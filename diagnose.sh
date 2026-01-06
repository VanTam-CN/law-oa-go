#!/bin/bash
echo "Diagnostics running" > diag.log
date >> diag.log
echo "Checking processes:" >> diag.log
ps aux | grep -E "main|app" >> diag.log
echo "Checking files:" >> diag.log
ls -l >> diag.log
