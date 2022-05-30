# rpm preuninstall script
# injected into %preun section
if [ $1 -eq "0" ] ; then
  systemctl stop hdl-srv-api
  systemctl disable hdl-srv-api
fi

exit 0