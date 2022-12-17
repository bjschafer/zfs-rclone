#!/usr/bin/env bash
# shellcheck disable=SC2154
set -x

#########################
# CONFIGURABLE OPTIONS #
#########################

PROCESSED_PROPERTY_NAME='backups:b2:last_processed'
SCHEDULE_PROPERTY_NAME='backups:b2:schedule'
RCLONE_OPTIONS=('--fast-list' '--progress' '--one-file-system') # '--syslog' 
RCLONE_REMOTE='b2_crypt'
RCLONE_STRATEGY='sync' # or copy, or move
ZFS_TYPE='filesystem' # or snap, or all, or vol
ZPOOL_NAME='warp'

####################
# INTERNAL OPTIONS #
####################

declare -A VALID_SCHEDULES
VALID_SCHEDULES[hourly]=$((60*60))
VALID_SCHEDULES[daily]=$((hourly*24))
VALID_SCHEDULES[weekly]=$((daily*7))
VALID_SCHEDULES[monthly]=$((weekly*4))

####################
# HELPER FUNCTIONS #
####################

should_process() {
    local schedule="$1"
    local last_processed="$2"

    local now
    now=$(date +%s)

    [[ $((now-last_processed)) -ge "${VALID_SCHEDULES[$schedule]}" ]]
}

exec_rclone() {
    local fsname="$1"

    echo "Running rclone:" "$RCLONE_STRATEGY" "${RCLONE_OPTIONS[@]}" "$fsname" "${RCLONE_REMOTE}:${fsname}"
    rclone "$RCLONE_STRATEGY" "${RCLONE_OPTIONS[@]}" "/${fsname}" "${RCLONE_REMOTE}:${fsname}"
}

#####################
# MAIN PROGRAM LOOP #
#####################

# get all candidates for processing
candidates=$(zfs get -H -o name,value -r "$SCHEDULE_PROPERTY_NAME" "$ZPOOL_NAME" -t "$ZFS_TYPE" | awk '$2 != "-"')

while read -r fsname schedule ; do
    if [[ ! -v "VALID_SCHEDULES[${schedule}]" ]]; then
        echo "Invalid schedule $schedule set on $fsname; skipping" >&2
        continue
    fi

    last_processed=$(zfs get -H -o value "$PROCESSED_PROPERTY_NAME" "$fsname")
    last_processed="${last_processed/-/0}"  # if not set, assume it's never been processed

    if should_process "$schedule" "$last_processed"; then
        exec_rclone "$fsname"
        zfs set "$PROCESSED_PROPERTY_NAME"="$(date +%s)" "$fsname"
    else
        echo "Skipping $fsname as it was last processed on $(date -d="$last_processed" +%c) and schedule is $schedule"
    fi

done <<< "$candidates"
