How to install collector command
=====================================

The collector command can run in a setup-hw container or on a bare metal Linux server.
This memo describes how to install the command on a bare metal server.

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

Put a bmc-user.json file in /etc/neco/ that must have "support" user to use the collector command.

```
{
  "support": {
    "password": {
      "raw": "raw password here"
    }
  }
}
```

Put a bmc-address.json file in /etc/neco. 

```
{
  "ipv4": {
    "address": "192.0.2.3",
    "netmask": "255.255.255.0",
    "gateway": "192.0.2.1"
  }
}
```

please see [config.md]("config.md") file.

## 4.Check 

By following the above steps, you can execute the collector command.

```
$ collector show
```

Next, see [how to generate rules](how_to_generate_rules.md).
