#!/bin/vbash

source /opt/vyatta/etc/functions/script-template
configure
/bin/cli-shell-api loadFile /config/config.boot
commit
save