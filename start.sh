#!/bin/bash
cd ./fin/ && bash startfin.sh $1 & cd ./tx_pool/ && bash start.sh $1
