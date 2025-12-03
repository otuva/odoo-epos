## Compile program for linux
```
go build -o /tmp/epos .
```
## Compile program for Windows x86_64
```
env GOOS=windows GOARCH=amd64 go build -o epos.exe
```
Then install epos.exe as a windows service by NSSM (https://nssm.cc/)

## Compile program for Raspberry Pi
```
env GOOS=linux GOARCH=arm GOARM=7 go build -o epos-armv7

> Note: The compiled 32-bit program (`GOARCH=arm GOARM=7`) can run on a 64-bit Raspberry Pi system as long as 32-bit compatibility is supported (most official Raspberry Pi OS versions support this by default).
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
Step 1.  If Odoo version <= 18.0, need to force your Point of Sale to use a secure connection (HTTPS)

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
    "usb": {
        "type": "usb",
        "address": "/dev/xp-n160ii",
        "paper_width": 576,
        "margin_bottom": 120
    },
    "serial": {
        "type": "serial",
        "address": "COM1,baud=115200,databits=8,parity=N,stopbits=1",
        "paper_width": 576,
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
        "type": "tcp",
        "address": "192.168.123.103:9100"
    },
    "png": {
        "type": "file",
        "address": "/tmp/pic"
    },
    "xp-236b": {
        "type": "usb",
        "address": "/dev/xp-236b"
    }
}
```