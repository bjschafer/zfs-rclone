# zfs-rclone

This script is designed to be dumped into `/etc/cron.hourly`. It uses [rclone](https://rclone.org/) to backup
ZFS datasets with a specific property on a periodic basis.

## Prerequisites

- Rclone configured with your desired remote for root (or whatever user you care to run this as).
- Sufficiently modern ZFS that supports user properties
- Modern `bash`

## Configuration

To configure backups for a ZFS dataset, run

```bash
zfs set backups:b2:schedule=daily my/dataset/name
```

Other supported schedule options are `hourly`, `daily`, `weekly`, and `monthly`.

All other config is handled via vars set at the top of the script:

| Variable                | Help                                                |
| --                      | -                                                   |
| PROCESSED_PROPERTY_NAME | ZFS property with timestmap of last successful sync |
| SCHEDULE_PROPERTY_NAME  | ZFS property to read for schedule                   |
| RCLONE_OPTIONS          | These flags are passed to `rclone`                  |
| RCLONE_REMOTE           | Name of already configured remote in `rclone`       |
| RCLONE_STRATEGY         | One of `sync`, `copy`, or `move`; see rclone docs   |
| ZFS_TYPE                | One of `filesystem`, `snap`, `vol`, or `all`        |
| ZPOOL_NAME              | Name of the base zpool to find datasets in          |

## Multiple remotes or zpools

If you require backing up to multiple rclone remotes, or from multiple zpools, you can run multiple instances of the script.
So long as `SCHEDULE_PROPERTY_NAME` and `PROCESSED_PROPERTY_NAME` are distinct for instances running on a zpool, there
should be no conflict.

Or, I'd happily welcome a PR to improve this :)

