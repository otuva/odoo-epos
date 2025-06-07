## Compile program for linux
```
go build -o /tmp/epos .
```
## Compile program for Windows x86_64
```
env GOOS=windows GOARCH=amd64 go build -o epos.exe
```
Then install epos.exe as a windows service by NSSM (https://nssm.cc/)

## Build deb package
```
go build -o linux/usr/local/odoo-epos/epos .
dpkg-deb --build linux epos.deb
```

## Features
* No need Odoo IOT Box
* No need ePOS printer
* Support Odoo online version
* No need any coding in Odoo
* Auto https certificate
* Widely compatiable with various 80mm thermal printers
* Can print receipt to png file

## How to use in Odoo
Step 1.  Force your Point of Sale to use a secure connection (HTTPS)

https://www.odoo.com/documentation/18.0/applications/sales/point_of_sale/configuration/https.html#force-your-point-of-sale-to-use-a-secure-connection-https

Step2. Install this program on Linux system

Step3. Edit your config.json file in /usr/local/odoo-epos

Step4. Restart this program:  sudo systemctl restart epos.service

Step5. Config the url in Odoo POS "ePOS Printer" Ip address field

Do NOT start with `https://`

```
localhost.ip.hogantech.net/p0
192-168-123-1.ip.hogantech.net/p1
```

## Sample of config.json
```
{
    "p0": {
        "type": "usb",
        "address": "/dev/xp-n160ii",
        "paper_width": 576,
        "margin_bottom": 120
    },
    "p1": {
        "type": "tcp",
        "address": "192.168.123.51:9100"
    },
    "p2": {
        "type": "tcp",
        "address": "192.168.123.52:9100"
    },
    "f0": {
        "type": "file",
        "address": "/tmp/receipts"
    }
}
```