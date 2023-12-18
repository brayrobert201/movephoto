#!/bin/bash

# Add execute permissions to the installer script
chmod +x "$0"

#!/bin/bash

# Add execute permissions to the installer script
chmod +x "$0"

# Create a systemd service file for movephoto
cat <<EOF | sudo tee /etc/systemd/system/movephoto.service
[Unit]
Description=Movephoto Service
After=network.target

[Service]
ExecStart=/path/to/movephoto --debug
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

# Start the movephoto service
sudo systemctl start movephoto.service

# Rest of your existing installer script goes here...
