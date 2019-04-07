Configuration Files
===================

`/etc/neco/bmc-address.json`
----------------------------

The file contains is a JSON object like this:

```json
{
    "ipv4": {
        "address": "1.2.3.4",
        "netmask": "255.255.255.0",
        "gateway": "1.2.3.1"
    }
}
```

BMC network interface will be configured to have the given `address`.


`/etc/neco/bmc-user.json`
-------------------------

This file contains credentials of BMC users.

BMC users are statically defined in `setup-hw` as follows:

* `root`: The administrator of BMC.
* `power`: Control power supply.
* `support`: Read-only account.

Credential types are:

* Raw password
* Hashed password with salt  
    For iDRAC, use [`idrac-passwd-hash`](../pkg/idrac-passwd-hash) tool to generate them.
* Authorized public keys for SSH

Supported credential types varies by BMC types.
iDRAC, BMC embedded in Dell servers, supports all credential types.

Example:

```json
{
    "root": {
        "password": {
            "raw": "raw password"
        },
        "authorized_keys": [
            "ssh-rsa ...",
            ...
        ]
    },
    "power": {
        "password": {
            "hash": "hashed_secret",
            "salt": "salt for hash"
        }
    }
}
```
