Configuration Files
===================

`/etc/neco/bmc-address.json`
----------------------------

The file contains a JSON object like this:

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
* `repair`: Restricted user who can execute repair operations, e.g. BMC reset.
* `power`: Restricted user who can control power supply.
* `support`: Read-only account.

All users are mandatory for iDRAC.

Credential types are:

* Raw password
* Hashed password with salt  
    For iDRAC, use [`idrac-passwd-hash`](../pkg/idrac-passwd-hash) tool to generate them.
    * `hash` and `salt` are for iDRAC9 and earlier.
    * `hash_salt` is for iDRAC10 (17G servers), which uses a different hash algorithm.
      17G servers require `raw` or `hash_salt`.

Supported credential types varies by BMC types.
iDRAC supports all credential types.

On 17G servers, the privileges of users other than `root` are assigned through
custom iDRAC roles named after the users (`repair`, `power` and `support`).
setup-hw creates the roles if they do not exist.

Example:

```json
{
    "root": {
        "password": {
            "raw": "raw password"
        }
    },
    "power": {
        "password": {
            "hash": "hashed_secret",
            "salt": "salt for hash",
            "hash_salt": "hashed_secret_for_idrac10|salt_for_idrac10"
        }
    }
}
```
