1. Copy `./systemd/go-ddns-client.service` to `/lib/systemd/system` and adjust the paths within this file
2. Start the service:

   ```bash
   sudo systemctl enable go-ddns-client
   sudo systemctl start go-ddns-client
   ```