# rpm postinstall script
# injected into %post section
systemctl daemon-reload &&
systemctl enable hdl-srv-api &&
systemctl restart hdl-srv-api

exit 0