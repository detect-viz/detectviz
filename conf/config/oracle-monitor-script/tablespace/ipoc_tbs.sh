#!/bin/bash
set -x
export ORACLE_HOME=/u01/app/oracle/product/19.3.0/dbhome_1 # need change
LogDir="/home/oracle/tbs"                                       # need change
DBFile="/home/oracle/tbs/account.csv"                           # need change
#################################################
########             set env             ########
#################################################
# crontab -e
# 0 * * * * /home/oracle/tbs/ipoc_db_tbs.sh
Host=`hostname`
Yesterday=$(perl -MPOSIX -le 'print strftime "%Y%m%d", localtime(time()-86400)')
DateTime=$(perl -MPOSIX -le 'print strftime "%Y%m%d%H%M", localtime(time())')
LogFile=`hostname`_$DateTime"_tableSpace.csv"
SentDir=$LogDir/SENT
BackupDir=$LogDir/BACKUP
FTPURL='51.15.225.188'
FTPUSER='bimap_test'
FTPPASS='#EDC4rfv'
RetentionDays=7
#################################################
########         collect command         ########
#################################################
while IFS=',' read -r username password sid
do
cd $LogDir
touch $LogFile
tmpFile=$LogDir"/output.txt"
$ORACLE_HOME/bin/sqlplus $username/$password@$sid <<EOF
set serveroutput on
@$LogDir/tbs.sql $tmpFile
EOF
cat $tmpFile >> $LogFile
done < $DBFile
#################################################
########    put awr file to ftpserver    ########
#################################################
Header="tablespace_name,autoextensible,files_in_tablespace,total_tablespace_space,total_used_space,total_tablespace_free_space,total_used_pct,total_free_pct,max_size_of_tablespace,max_free_size,total_auto_used_pct,total_auto_free_pct,date_time,sid,hostname"

if [ ! -d $SentDir ]; then mkdir $SentDir; fi
if [ ! -d $BackupDir ]; then mkdir $BackupDir; fi

rm -rf $tmpFile
sed -i '1 i '$Header $LogFile && mv $LogFile $SentDir


cd $SentDir
ftp -inv $FTPURL << EOF > $LogDir/ftplog.txt
user $FTPUSER $FTPPASS
cd oracle_bimaplnx06
binary
pwd

mput *
bye
EOF

# 如果FTP上傳成功，就移動至BACPUP檔案夾
msg=`grep "complete" $LogDir/ftplog.txt`
if [ -n "$msg" ]
    then
    mv $SentDir/*.csv $BackupDir
    echo "ftp upload complete"
else
    echo "ftp upload failed"
fi
rm -rf $LogDir/ftplog.txt

#################################################
########          zip and backup         ########
#################################################
# 1. 將昨日24個檔案壓縮為一個tar.gz
cd $BackupDir
if ls *$Yesterday*.csv 1> /dev/null 2>&1; then
    # 如果存在，將該檔案壓縮起來
    tar -czvf `hostname`"-"$Yesterday.tar.gz *$Yesterday*.csv && rm -rf *$Yesterday*.csv
fi

# 2. 根據RetentionDays定期將檔案清除
find $BackupDir/*.tar.gz -maxdepth 1 -mtime +$RetentionDays -type f -delete
exit
