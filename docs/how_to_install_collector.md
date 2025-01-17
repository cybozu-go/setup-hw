How to install collector command
=====================================

The collector command can run in setup-hw container and bare metal linux server. 
in this memo describe how to install bare metal server.

## 1.Clone code from github repository

```
$ git clone https://github.com/cybozu-go/setup-hw
```

## 2.Build & install collector command

```
$ cd setup-hw
$ make install
```

## 3.Setup config

Put a bmc-user.json file in /etc/neco/ that must have "support" user due to use the collector command.

```
{
  "support": {
    "password": {
      "raw": "cybozu"
    }
  }
}
```

Put a bmc-address.json file in /etc/neco. 

```
{"ipv4":{"address":"10.210.152.197","netmask":"255.255.252.0","maskbits":22,"gateway":"10.210.152.198"}}
```

please see [config.md]("config.md") file.

## 4.Check 

By following the above steps, you can execute the collector command.

```
$ collector show
```

Next, see [how to generate rules](how_to_install_collector.md).
