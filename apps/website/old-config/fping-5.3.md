# fping 5.3 離線版安裝說明（適用於 CentOS 7）

本執行檔為手動編譯的 **fping 5.3 for CentOS 7 x86_64** 版本，  
可直接複製至目標主機使用，無需安裝套件庫或相依套件。

---

## 檔案內容

- `fping`：主程式，可執行檔（編譯自 [fping.org](https://fping.org)）
- `README.md`：本說明檔

---

## 安裝步驟（建議用 root 權限）

### 1. 複製 fping 執行檔到系統路徑：

```bash
cp fping /usr/local/bin/fping
chmod +x /usr/local/bin/fping
```

---

### 2. 確保 `/usr/local/bin` 在 PATH 中：

檢查環境變數：

```bash
echo $PATH
```

若沒有 `/usr/local/bin`，請執行：

```bash
echo 'export PATH=$PATH:/usr/local/bin' > /etc/profile.d/fping-path.sh
chmod +x /etc/profile.d/fping-path.sh
source /etc/profile.d/fping-path.sh
```

---

### 3. 驗證安裝結果：

```bash
fping -v
# ➜ fping: Version 5.3
```

---

## 常見問題

| 問題                           | 解法                                      |
|--------------------------------|-------------------------------------------|
| `-bash: fping：命令找不到`    | 請確認 `/usr/local/bin` 是否在 `$PATH`，或重新開新 shell |
| 權限不足執行                  | 執行 `chmod +x /usr/local/bin/fping`     |
| 想要所有人都能使用            | 使用 root 安裝並確認 `/usr/local/bin` 路徑共享 |

---


## 範例用途

```bash
fping -g 192.168.1.1 192.168.1.254
fping -C 1 -q -g 10.1.249.1 10.1.249.254
```

---

## 來源與授權

本程式基於 fping 開源專案：  
- https://fping.org  
- 授權：BSD License

