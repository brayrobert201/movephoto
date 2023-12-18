#!/bin/bash

# Add execute permissions to the installer script
chmod +x "$0"

# Create a systemd service file for movephoto
cat <<EOF | sudo tee /etc/systemd/system/movephoto.service
[Unit]
Description=Movephoto Service
After=network.target

[Service]
ExecStart=/usr/local/bin/movephoto --watch
Restart=always
RestartSec=5
StandardOutput=syslog
StandardError=syslog
SyslogIdentifier=movephoto

[Install]
WantedBy=multi-user.target
EOF

# Reload systemd to recognize the new service
sudo systemctl daemon-reload

# Enable the movephoto service to start on boot
sudo systemctl enable movephoto.service

# Copy the movephoto executable and config.yaml.example to /usr/local/bin/
sudo cp ./movephoto /usr/local/bin/movephoto
sudo cp ./config.yaml.example /usr/local/bin/config.yaml
sudo chmod +x /usr/local/bin/movephoto

# Start the movephoto service
sudo systemctl start movephoto.service