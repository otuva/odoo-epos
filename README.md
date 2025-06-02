## Compile program
```
go build -o bin/epos .
```

## Build deb package
```
go build -o linux/usr/local/odoo-epos/epos .
dpkg-deb --build linux epos.deb
```