#!/bin/bash

bash changeconfig.sh && sleep 50 && bash launchinstances.sh $1 1 && sleep 70 && bash launchclient.sh 1 $2 && sleep 80 && mkdir $3 && bash scplog.sh $3 && sleep 20 && bash terminateinstances.sh
