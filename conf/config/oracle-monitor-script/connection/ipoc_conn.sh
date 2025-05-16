#!/bin/sh
#set env
#. /ub02/XXX/work/.profile

#!/bin/ksh
ORACLE_HOME=/oracle/EDV/112_64
COLT_HOME=/home/colt_dev
FTPURL='172.19.11.135'
FTPUSER='basis'
FTPPASS='cjo40'
Host=`hostname`
DateTime=$(perl -MPOSIX -le 'print strftime "%Y%m%d%H%M", localtime(time())')
LogFile=`hostname`_$DateTime"_oracleConn.err"
SentDir=$COLT_HOME/SENT
RetentionDays=7
AUTH_FILE=$COLT_HOME/account.csv
#################################################
########         collect command         ########
#################################################
cd $COLT_HOME
while IFS=',' read -r username password sid
do
    echo "exit" | $ORACLE_HOME/bin/sqlplus $username/$password@$sid | grep Connected > /dev/null
    if [ $? -eq 0 ] 
    then
        echo $sid "OK"
    else
        echo $sid "NOT OK"
        echo $(date +%s),${sid} >> $COLT_HOME/$LogFile
    fi
done < $AUTH_FILE

#################################################
########    put   file to ftpserver      ########
#################################################
if [ ! -d $SentDir ]; then mkdir $SentDir; fi
mv $COLT_HOME/$LogFile $SentDir
cd $SentDir
ftp -inv $FTPURL << EOF > $COLT_HOME/ftplog_conn.txt
user $FTPUSER $FTPPASS
binary
pwd
mput *
bye
EOF

msg=`grep "complete" $COLT_HOME/ftplog_conn.txt`
if [ -n "$msg" ]
    then
    rm -rf $SentDir/$LogFile
    echo "ftp upload complete"
else
    echo "ftp upload failed"
fi
rm $COLT_HOME/ftplog_conn.txt

exit
