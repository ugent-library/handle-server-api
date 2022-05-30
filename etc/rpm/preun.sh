# rpm preuninstall script
# injected into %preun section
if [ $1 -eq "0" ] ; then
  systemctl stop handle-server-api
  systemctl disable handle-server-api
fi

exit 0
