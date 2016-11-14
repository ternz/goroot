#!/bin/bash
pid_array=($(ps aux|grep "history_opt_service"|grep -v "grep" | awk -F " " '{print $2}'))

for pid in ${pid_array[@]}
do
{
	echo "kill  ${pid}"
	kill ${pid}
} &
done
wait


echo "stop finish!"
