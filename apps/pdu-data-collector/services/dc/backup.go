package dc

import (
	"bimap-zbox/databases"
	"bimap-zbox/global"
	"fmt"
	"io"
	"os"
	"os/exec"
	"syscall"
	"time"

	"go.uber.org/zap"
)

func RunBackupInfluxDB() error {
	event := global.Jobs.DCBackupInfluxDB
	startTime := time.Now() // 開始時間
	defer func() {
		// 結束時間
		duration := time.Since(startTime)
		
		tags := map[string]string{
			"name": event.Name,
		}
		fields := map[string]interface{}{
			"duration": int(duration.Seconds()),
			"taf":tags,
		}
		global.Logger.Info(event.Name, zap.Any("fields", fields), zap.Any("tags", tags))
		databases.WriteLogToInfluxDB(event.Name, tags, fields)

	}()
	db := global.EnvConfig.DC.InfluxDB
	tmpPath := db.BackupPath + "_tmp"

	cmd := exec.Command(db.InfluxExec, "backup", "--host", db.URL, "--token", db.Token, tmpPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("備份命令失敗: %v, %s", err, string(output))
	}

	// 備份成功後的處理：刪除舊的備份並更名新的備份資料夾
	err = cleanupBackup(tmpPath, db.BackupPath)
	if err != nil {
		return err
	}
	return nil
}

// 清理舊的備份資料夾
func cleanupBackup(tmpPath, bakPath string) error {
	// 刪除舊的備份資料夾
	if _, err := os.Stat(bakPath); err == nil {
		err = os.RemoveAll(bakPath)
		if err != nil {
			return fmt.Errorf("刪除舊的備份資料夾失敗: %v", err)
		}
	}
	// 更名新的備份資料夾
	if err := os.Rename(tmpPath, bakPath); err != nil {
		if linkErr, ok := err.(*os.LinkError); ok && linkErr.Err == syscall.EXDEV {
			// * 如果是跨裝置錯誤，使用複製和刪除的方法
			if err := copyAndDelete(tmpPath, bakPath); err != nil {
				global.Logger.Error(
					fmt.Sprintf("File copy and delete error: %v", err.Error()),
				)
				return err
			}
		} else {
			global.Logger.Error(
				fmt.Sprintf("File moved error: %v", err.Error()),
			)
			return linkErr
		}
	}
	fmt.Println("備份資料夾更名完成")
	return nil
}

// * 跨device則需使用此方法，但效能可能會比os.Rename略低，因為檔案複製比簡單的重命名操作需要更多的時間和資源
func copyAndDelete(src, dst string) error {
	// 開啟來源文件
	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer sourceFile.Close()

	// 建立目標文件
	destinationFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destinationFile.Close()

	// 複製文件內容
	_, err = io.Copy(destinationFile, sourceFile)
	if err != nil {
		return fmt.Errorf("failed to copy file content: %w", err)
	}

	// 確保所有資料都寫入目標文件
	err = destinationFile.Sync()
	if err != nil {
		return fmt.Errorf("failed to sync destination file: %w", err)
	}

	// 刪除來源文件
	err = os.Remove(src)
	if err != nil {
		return fmt.Errorf("failed to remove source file: %w", err)
	}

	return nil
}
