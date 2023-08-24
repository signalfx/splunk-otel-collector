#!/bin/sh
# Copyright Splunk, Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

# shellcheck disable=SC1000-SC9999 # Reason: This script is used in all the scripts and any change in this script would require a higher effort in testing all the scripts. Hence ignoring whole file.
# # # we don't want to point OS's utilities -- e.g. ntpdate(1) -- to libraries which Splunk bundles in SPLUNK_HOME/lib/
unset LD_PRELOAD LD_LIBRARY_PATH DYLD_LIBRARY_PATH SHLIB_PATH LIBPATH

# # # NIX-203 - set LANG env variable set to en_US to avoid parsing problems in other locales
EngLocale=`locale -a | grep -i "en_US.utf"`
if [ ! -z "$EngLocale" ]; then
    LANG=`echo $EngLocale | awk 'NR==1 {printf $1}'`
    export LANG
fi

# # # are we in debug mode?
if [ $# -ge 1 -a "x$1" = "x--debug" ] ; then
    DEBUG=1
    TEE_DEST=`dirname $0`/debug--`basename $0`--`date | sed 's/ /_/g;s/:/-/g'`
else
    DEBUG=0
    TEE_DEST=/dev/null
fi

DMESG_FILE=/var/log/dmesg
OS_FILE=/etc/os-release

# # # what OS is this?
KERNEL=`uname -s`
# # # what is the Kernel version?
KERNEL_RELEASE=`uname -r`

# # # assert we are in a supported OS
AWK=awk
case "x$KERNEL" in
    "xLinux")
        if [ -e $OS_FILE ]; then
            UBUNTU_MAJOR_VERSION=`awk -F'[".]' '/VERSION_ID=/ {print $2} ' $OS_FILE`;
        else
            UBUNTU_MAJOR_VERSION="";
            echo "$OS_FILE does not exist. UBUNTU_MAJOR_VERSION will be empty." > $TEE_DEST
        fi
        # # # enable check for OS versions, if needed later
        if [ -e /etc/debian_version ]; then DEBIAN=true; else DEBIAN=false; fi

        # # # /sbin/ is often absent in non-root users' PATH, and we want it for ifconfig(8)
        PATH=$PATH:/sbin/
        ;;
    "xSunOS")
        # # # enable check for OS versions, if needed later
        if [ `uname -r` = "5.8" ]; then SOLARIS_8=true; else SOLARIS_8=false; fi
        if [ `uname -r` = "5.9" ]; then SOLARIS_9=true; else SOLARIS_9=false; fi
        if [ `uname -r` = "5.10" ]; then SOLARIS_10=true; else SOLARIS_10=false; fi
        if [ `uname -r` = "5.11" ]; then SOLARIS_11=true; else SOLARIS_11=false; fi

        # # # eschew the antedeluvial awk
        AWK=nawk
        ;;
    "xDarwin")
        OSX_MINOR_VERSION=`sw_vers | sed -En '/ProductVersion/ s/^[^.]+\.([0-9]+)(\.[^.])?$/\1/p'`
        OSX_MAJOR_VERSION=`sw_vers | sed -En '/ProductVersion/ s/^[^0-9]+([0-9]+)\.[0-9]+(\.[^.]+)?$/\1/p'`

        # OSX_GE_SNOW_LEOPARD is for backward compatiblity.
        # Recommend that new code just use $OSX_MINOR_VERSION directly.
        if [ "$OSX_MAJOR_VERSION" == 10 ] && [ "$OSX_MINOR_VERSION" -ge 6 ]; then
            OSX_GE_SNOW_LEOPARD=true;
        else
            OSX_GE_SNOW_LEOPARD=false;
        fi

        ;;
    "xFreeBSD")
        ;;
    "xAIX")
        ;;
    "xHP-UX")
        ;;
    *)
        echo "UNIX flavor [$KERNEL] unsupported for Splunk *NIX App, quitting" > $TEE_DEST
        exit 1
        ;;
esac

# # # check for presence of required commands; we do not assume that which(1) exists, and roll our own
queryHaveCommand () # returns 0 if found, 1 if not
{
    [ "x$1" = "xeval" ] && shift
    for directory in `echo $PATH | sed 's/:/ /g'`
    do
        [ -x $directory/$1 ] && return 0
    done
    return 1
}

failLackCommand ()
{
    echo "Not found command [$1] on this host, quitting" > $TEE_DEST
    exit 1
}

failLackMultipleCommands ()
{
    echo "Not found any of commands [$*] on this host, quitting" > $TEE_DEST
    exit 1
}

assertHaveCommand ()
{
    queryHaveCommand $1
    if [ $? -eq 1 ] ; then
        failLackCommand $1
    fi
}

assertHaveCommandGivenPath ()
{
    [ "x$1" = "xeval" ] && shift
    [ -x $1 ] && return
    echo "Not found commandGivenPath [$1] on this host, quitting" > $TEE_DEST
    exit 1
}

failUnsupportedScript ()
{
    echo "UNIX flavor [$KERNEL] unsupported for this script, quitting" > $TEE_DEST
    exit 0
}

assertInvokerIsSuperuser ()
{
    [ `id -u` -eq 0 ] && return
    echo "Must be superuser to run this script, quitting" > $TEE_DEST
    exit 1
}

# # # check for presence of a few basic commands ubiquitous in our scripts
assertHaveCommand $AWK
assertHaveCommand egrep
