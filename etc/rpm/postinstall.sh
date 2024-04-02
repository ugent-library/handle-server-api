# rpm postinstall script
# injected into %post section
systemctl daemon-reload &&
systemctl enable handle-server-api &&
systemctl restart handle-server-api

exit 0
