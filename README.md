## Compile program
```
go build -o /tmp/epos .
```

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

```
https://localhost.ip.hogantech.net/p0
https://192-168-123-1.ip.hogantech.net/p1
```

## Sample of config.json
```
{
    "p0": {
        "type": "usb",
        "address": "/dev/usb/lp0",
        "margin_left": 16,
        "margin_bottom": 120
    },
    "p1": {
        "type": "tcp",
        "address": "192.168.123.101:9100"
    },
    "p2": {
        "type": "tcp",
        "address": "192.168.123.102:9100"
    },
    "p3": {
        "type": "file",
        "address": "/home/user/Pictures"
    }
}
```