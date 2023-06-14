#!/usr/bin/env bash
# Copyright The OpenTelemetry Authors
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

function build_scripted_input_endpoint()
# build a command name suitable for use in a REST target
{
    temp=`echo $1 | awk -F"/" '{print $NF}'`
    echo ".%252Fbin%252F"$temp
}

function build_monitor_input_endpoint()
# build a path name suitable for use in a REST target
{
    echo `echo $1 |  sed -e 's/\//%252F/g'`
}

function get_interval()
# get the given scripted input's interval
{
    interval=$(get_scripted_input_rest_value "$1" 'interval')
    echo $interval
}

function set_interval()
# set the given scripted input's interval
{
    set_scripted_input_rest_value "$1" "interval" "$2"
}

function set_metric_index()
# set the index for the given metric input
{
    set_scripted_input_rest_value "$1" "index" "$2"
}

function get_server_name
# get the server_name from 'show servername' cli
{
    if [ $remote_server_uri != "false" ]; then
        echo `$SPLUNK_HOME/bin/splunk show servername -uri $remote_server_uri | $AWK {'print $3'}`
    else
        echo `$SPLUNK_HOME/bin/splunk show servername | $AWK {'print $3'}`
    fi
}

function internal_call()
# low-level internal call handler
{
    if [ $remote_server_uri != "false" ]; then
        echo `$SPLUNK_HOME/bin/splunk _internal call /servicesNS/nobody/$remote_server_app_name/data/inputs/$1/$2 -uri $remote_server_uri`
    else
        echo `$SPLUNK_HOME/bin/splunk _internal call /servicesNS/nobody/$server_app_name/data/inputs/$1/$2`
    fi
}

function get_monitor_disabled_value()
{
    temp=$(internal_call 'monitor' "$1")
    for l in $temp; do
        case $l in
            *name=?disabled*) echo `echo $l | grep "name=\"disabled" | sed -e 's/^[[:space:]]*//' -e 's/[[:space:]]*$//' -e "s/name=\"disabled\">//" -e 's/<\/s:key>//g'`; break;;
        esac
    done
}

function get_monitor_status()
{
    echo "$input_counter) $1"
    input_endpoint=$(build_monitor_input_endpoint "$1")
    rest_value=$(get_monitor_disabled_value "$input_endpoint")
    case $rest_value in
        0)  echo "          enabled: *** disabled:     ";;
        1)  echo "          enabled:     disabled: *** ";;
    esac
}

function get_scripted_input_rest_value()
# given an scripted input endpoint and a key, set to $rest_value
{
    if [ $remote_server_uri != "false" ]; then
        echo `$SPLUNK_HOME/bin/splunk _internal call /servicesNS/nobody/$remote_server_app_name/data/inputs/script/$1 -uri $remote_server_uri | grep "name=\"$2" | sed -e 's/^[[:space:]]*//' -e 's/[[:space:]]*$//' -e "s/<s:key name=\"$2\">//" -e 's/<\/s:key>//g'`
    else
        echo `$SPLUNK_HOME/bin/splunk _internal call /servicesNS/nobody/$server_app_name/data/inputs/script/$1 | grep "name=\"$2" | sed -e 's/^[[:space:]]*//' -e 's/[[:space:]]*$//' -e "s/<s:key name=\"$2\">//" -e 's/<\/s:key>//g'`
    fi
}

function handle_rest_response()
# handle the rest response
{
    case $1 in
        *HTTP?Status:?200.*) echo "    $2 successful"; echo "";;
        *) echo "    $2 failed"; echo "";res="failure";;
    esac
}
function set_scripted_input_rest_value()
# given an endpoint and a post string, set the value
{
    setter_response=
    if [ $remote_server_uri != "false" ]; then
        setter_response=`$SPLUNK_HOME/bin/splunk _internal call /servicesNS/nobody/$remote_server_app_name/data/inputs/script/$1 -uri $remote_server_uri -post:$2 $3`
    else
        setter_response=`$SPLUNK_HOME/bin/splunk _internal call /servicesNS/nobody/$server_app_name/data/inputs/script/$1 -post:$2 $3`
    fi
    handle_rest_response "$setter_response" "update"
}

function enable_monitor_input()
# given a monitor input, enable it
{
    enable_response=
    if [ $remote_server_uri != "false" ]; then
        enable_response=`$SPLUNK_HOME/bin/splunk _internal call /servicesNS/nobody/$remote_server_app_name/data/inputs/monitor/$1/enable -uri $remote_server_uri -method POST`
    else
        enable_response=`$SPLUNK_HOME/bin/splunk _internal call /servicesNS/nobody/$server_app_name/data/inputs/monitor/$1/enable -method POST`
    fi
    handle_rest_response "$enable_response" "enable"
}

function disable_monitor_input()
# given a monitor input, disable it
{
    disable_response=
    if [ $remote_server_uri != "false" ]; then
        disable_response=`$SPLUNK_HOME/bin/splunk _internal call /servicesNS/nobody/$remote_server_app_name/data/inputs/monitor/$1/disable -uri $remote_server_uri -method POST`
    else
        disable_response=`$SPLUNK_HOME/bin/splunk _internal call /servicesNS/nobody/$server_app_name/data/inputs/monitor/$1/disable -method POST`
    fi
    handle_rest_response "$disable_response" "disable"
}
function enable_scripted_input()
# given a script name, enable it
{
    enable_response=
    if [ $remote_server_uri != "false" ]; then
        enable_response=`$SPLUNK_HOME/bin/splunk _internal call /servicesNS/nobody/$remote_server_app_name/data/inputs/script/$1/enable -uri $remote_server_uri -method POST`
    else
        enable_response=`$SPLUNK_HOME/bin/splunk _internal call /servicesNS/nobody/$server_app_name/data/inputs/script/$1/enable -method POST`
    fi
    handle_rest_response "$enable_response" "enable"
}

function disable_scripted_input()
# given a script name, disable it
{
    disable_response=
    if [ $remote_server_uri != "false" ]; then
        disable_response=`$SPLUNK_HOME/bin/splunk _internal call /servicesNS/nobody/$remote_server_app_name/data/inputs/script/$1/disable -uri $remote_server_uri -method POST`
    else
        disable_response=`$SPLUNK_HOME/bin/splunk _internal call /servicesNS/nobody/$server_app_name/data/inputs/script/$1/disable -method POST`
    fi
    handle_rest_response "$disable_response" "disable"
}

function update_app()
# updates the given app
{
    if [ $remote_server_uri != "false" ]; then
        install_response=`$SPLUNK_HOME/bin/splunk install app $1 -update true --uri $remote_server`
        case "$install_response" in
            *is?installed.* ) echo "    app install successful"; echo "";;
            *n?error?occurred:*) echo "    app install failed"; echo "";;
        esac
    else
        install_response=`$SPLUNK_HOME/bin/splunk install app $1 -update true`
        case "$install_response" in
            *is?installed.* ) echo "    app install successful"; echo "";;
            *n?error?occurred:*) echo "    app install failed"; echo "";;
        esac
    fi
}

function install_app()
# installs the app residing at the given remote path
{
    if [ $remote_server_uri != "false" ]; then
        install_response=`$SPLUNK_HOME/bin/splunk install app $1 -uri $remote_server_uri`
        case "$install_response" in
            *is?installed.* ) echo "    app install successful"; echo "";;
            *install?anywa* ) echo "    app already installed.  Attempting to upgrade"; update_app "$1";;
            *n?error?occurred:*) echo "    app install failed - the URI provided was not found"; echo "";;
            * ) echo "ERROR: $install_response";;
        esac
    else
        install_response=`$SPLUNK_HOME/bin/splunk install app $1`
        case "$install_response" in
            *is?installed.* ) echo "    app install successful"; echo "";;
            *install?anywa* ) echo "    app already installed.  Attempting to upgrade"; update_app "$1";;
            *n?error?occurred:*) echo "    app install failed - the URI provided was not found"; echo "";;
            * ) echo "ERROR: $install_response";;
        esac
    fi
}

function get_scripted_input_status()
# given an input, get the enabled/disabled
# status and, if enabled, the interval
{
    echo "$input_counter) $1"
    input_endpoint=$(build_scripted_input_endpoint "$1")
    rest_value=$(get_scripted_input_rest_value "$input_endpoint" 'disabled')
    index_value=$(get_scripted_input_rest_value "$input_endpoint" 'index')
    if [ "$rest_value" = "0" ]; then
        interval=$(get_interval "$input_endpoint")
        if [ "$interval" != "false" ]; then
            echo "          enabled: *** disabled:      interval: $interval      index: $index_value"
        else
            echo "          enabled: *** disabled:      index: $index_value"
        fi

    else
        echo "          enabled:     disabled: ***      index: $index_value"
    fi
}

function get_script_list
# sets the scripted input list in $output
{
    if [ $remote_server_uri != "false" ]; then
       echo `$SPLUNK_HOME/bin/splunk list exec -uri "$remote_server_uri"`
    else
       echo `$SPLUNK_HOME/bin/splunk list exec`
    fi
}

function show_inputs
# show input status parsed from 'list exec'
# if enabled show the interval and last run time
{
    clear
    echo ""
    echo "*** Splunk> *nix command-line setup > SHOW INPUT STATUS ***"
    echo ""
    input_counter=0
    echo " Scripted Inputs:"
    echo ""
    script_list=$(get_script_list)
    for line in $script_list; do
        case "$line" in
           *unix* | *Splunk_TA_nix* ) get_scripted_input_status "$line"; input_counter=`expr $input_counter + 1`;
        esac
    done
    echo ""
    echo " Monitor Inputs:"
    echo ""
    for line in $MONITOR_INPUTS; do
        get_monitor_status "$line"
        input_counter=`expr $input_counter + 1`
    done
}

function enable_all_inputs
#enables all endpoints
{
    oldIFS=$IFS
    IFS='
    '
    script_list=$(get_script_list)
    for line in $script_list; do
        res="success"
        flag=0
        if [[ $line == *"_metric"* && ! -z $1 ]]; then
            input_endpoint=$(build_scripted_input_endpoint "$line")
            echo "updating index of $line to $1"
            set_metric_index "$input_endpoint" "$1"
            flag=1
        fi
        if [ "$res" == "success" ] && [[ ( $line != *"_metric"* || $flag == 1 ) ]]; then
            case "$line" in
            *unix* | *Splunk_TA_nix* ) echo "enabling $line"; input_endpoint=$(build_scripted_input_endpoint "$line"); enable_scripted_input $input_endpoint;;
            esac
        fi
    done
    for line in $MONITOR_INPUTS; do
        echo "enabling $line"
        input_endpoint=$(build_monitor_input_endpoint "$line")
        enable_monitor_input $input_endpoint
    done
    IFS=$oldIFS
    echo ""
}

function disable_all_inputs
# disables all inputs
{
    #oldIFS=$IFS
    #IFS='
    #'
    script_list=$(get_script_list)
    for line in $script_list; do
        case "$line" in
           *unix* | *Splunk_TA_nix* ) echo "disabling $line"; input_endpoint=$(build_scripted_input_endpoint "$line"); disable_scripted_input $input_endpoint;;
        esac
    done
    for line in $MONITOR_INPUTS; do
        echo "disabling $line"
        input_endpoint=$(build_monitor_input_endpoint "$line")
        disable_monitor_input "$input_endpoint"
    done
    #IFS=$oldIFS
    echo ""
}

function set_remote_input()
# set the given configuration on the remote host
{
    _input_type=
    _input=
    _disabled=
    for value in $1; do
        if [ ! -n "$_input_type" ]; then
            _input_type="$value"
        else
            if [ "$_input_type" == "monitor" ]; then
                if [ ! -n "$_input" ]; then
                    _input="$value"
                else
                    if [ "$value" == "1" ]; then
                        disable_monitor_input "$_input"
                    else
                        enable_monitor_input "$_input"
                    fi
                fi
            else
                if [ ! -n "$_input" ]; then
                    _input="$value"
                else
                    if [ ! -n "$_disabled" ]; then
                        _disabled="$value"
                    else
                        if [ "$_disabled" == "1" ]; then
                            disable_scripted_input "$_input"
                        else
                            enable_scripted_input "$_input"
                            set_interval "$_input" "$value"
                        fi
                    fi
                fi
            fi
        fi
    done
}

function monitor_clone()
# clone monitor input
{
    _remote_server_uri=$remote_server_uri
    remote_server_uri="false"
    input_endpoint=$(build_monitor_input_endpoint "$1")
    rest_value=$(get_monitor_disabled_value "$input_endpoint")
    remote_server_uri=$_remote_server_uri
    set_remote_input "monitor $input_endpoint $rest_value"
}

function scripted_clone()
# clone scripted input
{
    interval=
    _remote_server_uri=$remote_server_uri
    remote_server_uri="false"
    input_endpoint=$(build_scripted_input_endpoint "$1")
    rest_value=$(get_scripted_input_rest_value "$input_endpoint" 'disabled')
    remote_server_uri=$_remote_server_uri
    if [ "$rest_value" = "0" ]; then
        interval=$(get_interval "$input_endpoint")
        set_remote_input "scripted $input_endpoint $rest_value $interval"
    else
        set_remote_input "scripted $input_endpoint $rest_value"
    fi
}

function clone_all_inputs
# clone all inputs from local to remote_server_uri
{
    if [ $_remote_server_uri == "false" ]; then
        echo ""
        echo "    No remote server is set"
        echo ""
        echo "    Please specify a remote server through the main menu"
        echo "    or via command line arguments in order to clone inputs"
        echo ""
    else
        echo ""
        echo "    copying local input configuration to $server_name"
        echo ""
        echo "    Please be patient, this might take a minute..."
        echo ""
        script_list=$(get_script_list)
        for line in $script_list; do
            case "$line" in
                *unix* | *Splunk_TA_nix* ) echo ""; echo "    cloning $line to $server_name"; echo ""; scripted_clone "$line"
            esac
        done
        for line in $MONITOR_INPUTS; do
            echo ""
            echo "    cloning $line to $server_name"
            echo ""
            monitor_clone "$line"
        done
    fi
}

function enable_all_menu
# batch enable all inputs
{
    clear
    echo ""
    echo "*** Splunk> *nix command-line setup > ENABLE ALL INPUTS ***"
    echo ""
    echo "You are currently managing Splunk server '$server_name'"
    echo ""
    echo "1 - confirm and enable all inputs"
    echo "2 - return to the manage inputs menu"
    echo ""
    read selection
    echo ""

    case $selection in
        1 ) echo "";echo "Do you want to enable metric inputs too, if yes, enter metric index name else press enter";read metric_index;if [ ! -z $metric_index ]; then enable_all_inputs "$metric_index"; else enable_all_inputs; fi; press_enter;manage_inputs_menu;;
        2 ) manage_inputs_menu;;
        * ) echo "Please enter a number between 1 and 2"; press_enter; enable_all_menu;;
    esac
}

function disable_all_menu
# batch disable all inputs
{
    clear
    echo ""
    echo "*** Splunk> *nix command-line setup > DISABLE ALL INPUTS ***"
    echo ""
    echo "You are currently managing Splunk server '$server_name'"
    echo ""
    echo "1 - confirm and disable all inputs"
    echo "2 - return to the manage inputs menu"
    echo ""
    echo -n "Please enter your selection: "
    read selection
    echo ""
    case $selection in
        1 ) disable_all_inputs; press_enter; manage_inputs_menu;;
        2 ) manage_inputs_menu;;
        * ) echo "Please enter a number between 1 and 2"; press_enter; disable_all_menu;;
    esac
}

function local_to_remote_menu
# confirm local to remote config copy
{
    clear
    echo ""
    echo "*** Splunk> *nix command-line setup > COPY LOCAL CONFIG TO REMOTE ***"
    echo ""
    echo "You are currently managing Splunk server '$server_name'"
    echo ""
    echo "1 - confirm and clone all local inputs to $server_name"
    echo "2 - return to the manage inputs menu"
    echo ""
    echo -n "Please enter your selection: "
    read selection
    echo ""
    case $selection in
        1 ) clone_all_inputs; press_enter; manage_inputs_menu;;
        2 ) manage_inputs_menu;;
        * ) echo "Please enter a number between 1 and 2"; press_enter; local_to_remote_menu;;
    esac
}

function change_input_interval()
# change the input's interval
{
    echo ""
    echo ""
    echo -n "Enter the new interval value: "
    read selection
    echo ""
    if test $selection -ge 0; then
        input_endpoint=$(build_scripted_input_endpoint "$1")
        set_interval "$input_endpoint" "$selection"
    else
        echo ""
        echo "The value you entered is not a number - please try again"
        echo ""
        change_input_interval $1
    fi

}

function toggle_scripted_input()
# handle enable/disable of scripted input
{
    if [ "$2" = "0" ]; then
        input_endpoint=$(build_scripted_input_endpoint "$1")
        disable_scripted_input "$input_endpoint"
    else
        input_endpoint=$(build_scripted_input_endpoint "$1")
        enable_scripted_input "$input_endpoint"
    fi
}

function toggle_monitor_input()
# handle enable/disable of monitor input
{
    if [ "$2" = "0" ]; then
        input_endpoint=$(build_monitor_input_endpoint "$1")
        disable_monitor_input "$input_endpoint"
    else
        input_endpoint=$(build_monitor_input_endpoint "$1")
        enable_monitor_input "$input_endpoint"
    fi

}

function manage_scripted_input_options()
# show scripted input settings/options and handle input
{
    get_scripted_input_status "$1"
    echo ""
    echo "    Please choose from one of the following options:"
    echo ""
    if [ "$rest_value" = "0" ]; then
        echo "1 - disable input"
    else
        echo "1 - enable input"
    fi
    echo "2 - change input interval"
    echo "3 - return to the previous menu"
    echo ""
    echo "0 - logout and exit program"
    echo ""
    echo -n "Please enter your selection: "
    read selection
    echo ""
    case $selection in
        1) toggle_scripted_input "$1" "$rest_value"; press_enter; manage_input_menu "$1";;
        2) change_input_interval "$1"; press_enter; manage_input_menu "$1";;
        3) select_input_menu;;
        0) splunk_logout; exit 0;;
        *) echo "please enter a number between 0 and 3"; manage_input_menu "$1";;
    esac
}

function manage_monitor_input_options()
# show monitor input settings/options and handle input
{
    get_monitor_status "$1"
    echo ""
    echo "    Please choose from one of the following options:"
    echo ""
    if [ "$rest_value" = "0" ]; then
        echo "1 - disable input"
    else
        echo "1 - enable input"
    fi
    echo "2 - return to the previous menu"
    echo ""
    echo "0 - logout and exit program"
    echo ""
    echo -n "Please enter your selection: "
    read selection
    echo ""
    case $selection in
        1) toggle_monitor_input "$1" "$rest_value"; press_enter; manage_input_menu "$1";;
        2) select_input_menu;;
        0) splunk_logout; exit 0;;
        *) echo "please enter a number between 0 and 2"; manage_input_menu "$1";;
    esac
}

function manage_input_menu()
# manage one input
{
    clear
    echo ""
    echo "*** Splunk> *nix command-line setup > CHOOSE INPUT TO MANAGE ***"
    echo ""
    echo "You are currently managing Splunk server '$server_name'"
    echo ""
    echo "--> Manage Input '$1'"
    echo ""
    res="success"
    input_endpoint=$(build_scripted_input_endpoint "$1")
    rest_index=$(get_scripted_input_rest_value "$input_endpoint" 'index')
    if [[ "$1" == *"_metric"* ]] ; then
        if [[ "$rest_index" != "default" ]]; then
            echo "Do you want to change the metric index (y/n)?"
            read answer

            if [[ "$answer" == "y" ]]; then
                echo "Enter the metric index"
                read metric_index
                if [ ! -z $metric_index ]; then
                    input_endpoint=$(build_scripted_input_endpoint "$1")
                    set_metric_index $input_endpoint $metric_index
                else
                    echo "Please enter a valid index"
                    press_enter
                    manage_input_menu "$1"
                fi
            fi
        else
            echo "Enter the metric index"
            read metric_index
            if [ ! -z $metric_index ]; then
                input_endpoint=$(build_scripted_input_endpoint "$1")
                set_metric_index $input_endpoint $metric_index
            else
                echo "Please enter a valid index"
                press_enter
                manage_input_menu "$1"
            fi
        fi
    fi
    if [ $res == "success" ]; then
        case "$1" in
            *.sh) manage_scripted_input_options $1;;
            *) manage_monitor_input_options $1;;
        esac
    else
        press_enter
        select_input_menu
    fi
}

function select_input_menu
# choose one input, then enable/disable/change interval
{
    clear
    echo ""
    echo "*** Splunk> *nix command-line setup > CHOOSE INPUT TO MANAGE ***"
    echo ""
    echo "You are currently managing Splunk server '$server_name'"
    echo ""
    echo ""
    echo "  Choose one of the following inputs:"
    echo ""
    selection_list=()
    input_counter=1
    oldIFS=$IFS
    IFS='
    '
    script_list=$(get_script_list)
    for line in $script_list; do
        case "$line" in
           *unix* | *Splunk_TA_nix* ) echo " $input_counter - $line"; selection_list[$input_counter]=$line; input_counter=`expr $input_counter + 1`;
        esac
    done
    for line in $MONITOR_INPUTS; do
        echo " $input_counter - $line"
        selection_list[$input_counter]=$line
        input_counter=`expr $input_counter + 1`
    done
    echo ""
    echo " $input_counter - go back to manage inputs menu"
    echo ""
    echo ""
    echo "  0 - logout and exit program"
    echo ""
    echo -n "Enter selection: "
    read selection
    echo ""
    if [ $selection = $input_counter ]; then
        manage_inputs_menu
    elif [ $selection = 0 ]; then
        splunk_logout
        exit 0
    elif [ $selection -gt $input_counter ]; then
        echo "Please enter a number between 0 and $input_counter"
        press_enter
        select_input_menu
    elif [ $selection -lt 0 ]; then
        echo "Please enter a number between 0 and $input_counter"
        press_enter
        select_input_menu
    else
        ### TODO: implement manage_selected_input_menu
        manage_input_menu ${selection_list[$selection]}
    fi
}

function manage_inputs_menu
# the aptly named 'manage inputs' menu
{
    clear
    echo ""
    echo "*** Splunk> *nix command-line setup > MANAGE INPUTS ***"
    echo ""
    echo "You are currently managing Splunk server '$server_name'"
    echo ""
    echo "    Please choose from one of the following options:"
    echo ""
    echo "1 - manage one input"
    echo "2 - enable all inputs"
    echo "3 - disable all inputs"
    if [ "$remote_server_uri" != "false" ] && [ "$server_unix_app_installed" = "true" ]; then
        echo "4 - copy local configuration to remote"
        echo "5 - go back to main menu"
        echo ""
        echo "0 - logout and exit program"
        echo ""
        echo -n "Enter selection: "
        read selection
        echo ""
        case $selection in
            1 ) select_input_menu;;
            2 ) enable_all_menu;;
            3 ) disable_all_menu;;
            4 ) local_to_remote_menu;;
            5 ) main_menu ;;
            0 ) splunk_logout; exit 0 ;;
            * ) echo "Please enter a number between 0 and 4"; press_enter; manage_inputs_menu;;
        esac
    else
        echo "4 - go back to main menu"
        echo ""
        echo "0 - logout and exit program"
        echo ""
        echo -n "Enter selection: "
        read selection
        echo ""
        case $selection in
            1 ) select_input_menu;;
            2 ) enable_all_menu;;
            3 ) disable_all_menu;;
            4 ) main_menu ;;
            0 ) splunk_logout; exit 0 ;;
            * ) echo "Please enter a number between 0 and 4"; press_enter; manage_inputs_menu;;
        esac
    fi
}

function install_menu
# the aptly named install menu
{
    clear
    echo ""
    echo "*** Splunk> *nix command-line setup > INSTALL/UPGRADE MENU***"
    echo ""
    echo "You are currently managing Splunk server '$server_name'"
    echo ""
    echo " Please enter the full URI string indicating where the app resides"
    echo ""
    echo "   -> for example, 'https://localhost/apps/unix_app_new.tgz'"
    echo ""
    echo -n "Enter URI: "
    read install_uri
    install_app "$install_uri"
    press_enter
    main_menu
}

function press_enter
# convenience function to prompt for return
{
    echo ""
    echo -n "Press Enter to continue"
    read
    clear
}

function main_menu
# the aptly named main menu
{
    clear
    echo ""
    echo "*** Splunk> *nix command-line setup > MAIN MENU ***"
    echo ""
    echo "You are currently managing Splunk server '$server_name'"
    echo ""
    echo "    Please choose from one of the following options:"
    echo ""
    echo "1 - show *nix input status"
    echo "2 - manage *nix inputs"
    echo "3 - install/upgrade app"
    echo "4 - change credentials"
    if [ $remote_server_uri != "false" ]; then
        echo "5 - disconnect from remote instance"
    else
        echo "5 - connect to remote instance"
    fi
    echo ""
    echo "0 - logout and exit program"
    echo ""
    echo -n "Enter selection: "
    read selection
    echo ""
    case $selection in
        1 ) show_inputs; press_enter; main_menu ;;
        2 ) manage_inputs_menu;;
        3 ) install_menu;;
        4 ) handle_credential_change;;
        5 ) handle_remote_connection;;
        0 ) splunk_logout; exit 0;;
        * ) echo "Please enter a number between 0 and 5"; press_enter; main_menu;;
    esac
}

function set_app_installed()
# set the appropriate remote or local app installed flag
{
    if [ $remote_server_uri != "false" ]; then
        remote_server_unix_app_installed="true"
        remote_server_app_name="$1"
    else
        server_unix_app_installed="true"
        server_app_name="$1"
    fi
}

function set_app_enabled
# if app is enabled, set the appropriate variables
{
    if [ $remote_server_uri != "false" ]; then
        if [ $remote_server_unix_app_installed != "false" ]; then
            set_server_has_app_enabled
        else
            unset_server_has_app_enabled
        fi
    else
        if [ $server_unix_app_installed != "false" ]; then
            set_server_has_app_enabled
        else
            unset_server_has_app_enabled
        fi
    fi
}

function set_server_has_app_enabled
# set appropriate flag that server has
# the unix app installed and enabled
{
    if [ $remote_server_uri != "false" ]; then
        remote_server_has_unix_app_enabled="true"
    else
        server_has_unix_app_enabled="true"
    fi
}

function unset_server_has_app_enabled
# set appropriate flag that server does not
# have the unix app installed and enabled
{
    if [ $remote_server_uri != "false" ]; then
        remote_server_has_unix_app_enabled="false"
    else
        server_has_unix_app_enabled="false"
    fi
}

function handle_credential_change
# handle remote or local credential change
{
    if [ $remote_server_uri != "false" ]; then
        splunk_remote_credential_change
    else
        splunk_logout
        splunk_login
    fi
}

function handle_remote_connection
# if connected to remote instance, logout
# else redirect to remote instance login
{
    if [ $remote_server_uri != "false" ]; then
        splunk_remote_logout
    else
        splunk_remote_login
    fi
}

function set_unix_app_info
{
    if [ $remote_server_uri != "false" ]; then
        app_output=`$SPLUNK_HOME/bin/splunk display app -uri $remote_server_uri`
    else
        app_output=`$SPLUNK_HOME/bin/splunk display app`
    fi
    oldIFS=$IFS
    IFS='
    '
    for line in $app_output; do
        case "$line" in
           *unix* ) set_app_installed "unix";;
           *Splunk_TA_nix* ) set_app_installed "Splunk_TA_nix";;
           *ENABLED*) set_app_enabled;;
           #*DISABLED*) set_app_disabled;;
        esac
    done
    IFS=$oldIFS
}

function check_for_unix_app
# can't manage the unix app if there is nothing to manage
{
    set_unix_app_info
    if [ $remote_server_uri = "true" ]; then
        if [ $remote_server_has_unix_app_enabled = "true" ]; then
            main_menu
        else
            echo "the remote server $server_name does not have the unix app installed or the app is disabled"
            echo ""
            echo "do you want to install the unix app from a location on your network?"
            echo ""
            echo -n "enter y to continue: "
            read want_install_app
            case $want_install_app in
                y ) install_menu; check_for_unix_app;;
                * ) splunk_remote_logout; prerequisites;;
            esac
        fi
    else
        if [ $server_has_unix_app_enabled = "true" ]; then
            main_menu
        else
           echo "the local server $server_name does not have the unix app installed or the app is disabled"
           echo ""
           echo "only remote management of servers with the unix app will be permitted"
           splunk_remote_login
        fi
    fi
}

function prerequisites
# use 'list app' to see if the unix app is installed/enabled
# set server_name
# if app installed/enabled, redirect to main menu
# else warn and exit
{
    server_name=$(get_server_name)
    check_for_unix_app
    main_menu
}

function splunk_login
# log user in to splunk
# then route to main_menu
{
    clear
    echo ""
    echo "*** Splunk> *nix command-line setup > LOCAL LOGIN ***"
    echo ""
    $SPLUNK_HOME/bin/splunk login
    if [ "$?" = "0" ]; then
        prerequisites
    else
        exit 1
    fi
}

function splunk_remote_login
# log user in to some other splunk
# then route to main_menu
{
    clear
    echo ""
    echo "*** Splunk> *nix command-line setup > REMOTE LOGIN ***"
    echo ""
    echo " Please enter the full URI for the remote server"
    echo ""
    echo "   -> for example, 'https://remotehost:8089'"
    echo ""
    echo -n "Enter URI: "
    read remote_server_uri
    splunk_remote_credential_change
}

function splunk_remote_credential_change
# branch the remote credential change to facilitate
# changing credentials on the same remote instance
{
    echo ""
    echo "connecting to the remote server '$remote_server_uri'"
    echo ""
    echo "enter your credentials to the remote server below:"
    echo ""
    $SPLUNK_HOME/bin/splunk login --uri "$remote_server_uri"
    if [ "$?" = "0" ]; then
        prerequisites
    else
        remote_server_uri="false"
        remote_server_unix_app_installed="false"
        remote_server_has_unix_app_enabled="false"
        echo ""
        echo "remote login failed"
        echo ""
        press_enter
        main_menu
    fi
}

function splunk_logout
# log user out of splunk
# often followed by call to splunk_login
{
    $SPLUNK_HOME/bin/splunk logout
    remote_server_uri="false"
    server_name="false"
    server_unix_app_installed="false"
    server_has_unix_app_enabled="false"
    remote_server_unix_app_installed="false"
    remote_server_has_unix_app_enabled="false"
    clear
}

function splunk_remote_logout
# log user out of remote splunk instance
{
    $SPLUNK_HOME/bin/splunk logout --uri "$remote_server_uri"
    remote_server_uri="false"
    remote_server_unix_app_installed="false"
    remote_server_has_unix_app_enabled="false"
    splunk_login
    server_name=$(get_server_name)
    main_menu
}

function usage()
# provides usage
{
    echo ''
    echo '  usage: setup.sh'
    echo ''
    echo '       (no argument)   menu-based setup'
    echo '       --auth          credentials (user:pass) for specified command'
    echo '       --clone-all     clone input configuration from local to remote'
    echo '       --disable-all   disable all inputs'
    echo '       --disable-input input to be disabled'
    echo '       --enable-all    enable all inputs. Metric inputs will be enabled if metric input will be passed'
    echo '       --enable-input  input to be enabled and metric index must be passed for metric input'
    echo '       --help          print usage and exit'
    echo '       --install-app   install the app at the given location'
    echo '       --interval      set input to given interval'
    echo '       --list-all      show details all inputs'
    echo '       --list-input    show details for input'
    echo '       --usage         print usage and exit'
    echo '       --uri           remote uri (https://host:port) to use'
    echo '       --metric-index  provide metric index in metric input'
    echo ''
    echo ''
    echo '  examples:'
    echo ''
    echo '       set cpu.sh interval to 120 (with auth prompt):'
    echo ''
    echo '           setup.sh --interval cpu.sh 120'
    echo ''
    echo '       disable all local inputs (with no auth prompt):'
    echo ''
    echo '           setup.sh --disable-all --auth admin:changeme1'
    echo ''
    echo '       show input status on remote host foobar:'
    echo ''
    echo '           setup.sh --list-all --uri https://foobar:8089'
    echo ''
    echo '       update the unix app from your-server on the remote host foobar:'
    echo ''
    echo '           setup.sh --install-app https://your-server/unix.spl --uri https://foobar:8089'
    echo ''
    echo '       copy the local input configuration to the remote host foobar:'
    echo ''
    echo '           setup.sh --clone-all --uri https://foobar:8089'
    echo ''
    echo '       enable all inputs including metric inputs'
    echo ''
    echo '           setup.sh --enable-all --metric-index test3'
    echo ''
    echo '       enable a single metric input'
    echo ''
    echo '           setup.sh --enable-input interfaces_metric.sh --metric-index test3'
    echo ''

    exit 1
}

function execute_command()
# executes one command from the execution queue
{
    action=
    _target=
    _interval=
    res="success"
    for token in $1; do
        if [ ! -n "$action" ]; then
            action="$token"
            continue
        else
            if [ "$action" == "clone" ]; then
                clone_all_inputs
            elif [ "$action" == "disable" ]; then
                if [ "$token" == "all" ]; then
                    disable_all_inputs
                else
                    case $token in
                        *.sh ) input_endpoint=$(build_scripted_input_endpoint "$token"); echo "disabling input $token"; echo ""; disable_scripted_input "$input_endpoint";;
                        * ) input_endpoint=$(build_monitor_input_endpoint "$token"); echo "disabling input $token"; echo ""; disable_monitor_input "$input_endpoint";;
                    esac
                fi
            elif [ "$action" == "enable" ]; then
                word=( $1 )
                if [ "$token" == "all" ]; then
                    if [ ${#word[@]} == "2" ] || [ ${#word[@]} == "3" ]; then
                        echo ""
                        echo "Warning <<<<<<<<< Metric inputs will not be enabled as metric index was not specified >>>>>>>>>"
                        echo ""
                        enable_all_inputs
                    elif [ ${#word[@]} == "4" ]; then
                        if [ "${word[2]}" == "--metric-index" ]; then
                            enable_all_inputs ${word[3]}
                        else
                            echo "Wrong Argument"
                            usage
                        fi
                    else
                        echo "Wrong argument"
                        usage
                    fi
                elif [ "$token" == "input" ]; then
                    _target=${word[2]}
                    if [ ${#word[@]} == "3" ] ; then
                        if [[ "$_target" != *"_metric"* ]]; then
                            enable_single_input $_target
                        else
                            echo "Metric index must be specified for this input"
                            usage
                        fi
                    elif  [ ${#word[@]} == "4" ] ; then
                        echo "Wrong argument"
                        usage
                    elif [ ${#word[@]} == "5" ]; then
                        if [[ "${word[3]}" == "--metric-index" ]] && [[ "$_target" == *"_metric"* ]]; then
                            enable_metric_input $_target ${word[4]}
                        else
                            echo "This input is not a metric input or wrong argument passed"
                            usage
                        fi
                    else
                        echo "Wrong Argument"
                        usage
                    fi
                fi
            elif [ "$action" == "install" ]; then
                install_app "$token"
            elif [ "$action" == "interval" ]; then
                if [ ! -n "$_target" ]; then
                    _target="$token"
                else
                    if [ ! -n "$_interval" ]; then
                        input_endpoint=$(build_scripted_input_endpoint "$_target")
                        echo "setting $_target interval to $token"
                        set_interval "$input_endpoint" "$token"
                    fi
                fi
            elif [ "$action" == "list" ]; then
                if [ "$token" == "all" ]; then
                    show_inputs
                else
                    case "$token" in
                        *.sh ) input_endpoint=$(build_scripted_input_endpoint "$token"); get_scripted_input_status "$input_endpoint";;
                        * ) input_endpoint=$(build_monitor_input_endpoint "$token"); get_monitor_status "$input_endpoint";;
                    esac
                fi
            fi
        fi
    done
    }

function enable_metric_input
# Updates index of metric input and if successful then enable it.
{
    input_endpoint=$(build_scripted_input_endpoint "$1")
    set_metric_index "$input_endpoint" "$2"
    if [ "$res" == "success" ]; then
        enable_single_input "$1"
    fi
}

function enable_single_input
# Enable any input
{
    case $1 in
        *.sh ) input_endpoint=$(build_scripted_input_endpoint "$1"); echo "enabling input $1"; echo ""; enable_scripted_input "$input_endpoint";;
        * ) input_endpoint=$(build_monitor_input_endpoint "$1"); echo "enabling input $1"; echo ""; enable_monitor_input "$input_endpoint";;
    esac
}

function execute_queue
# executes a stored queue of command line options and arguments
{
    if [ ! -n "$__QUEUE" ]; then
        echo ""
        echo " Error parsing command line options/arguments"
        echo ""
        echo ""
        usage
    else
        if [ -n "$AUTH_STRING" ]; then
            if [ "$remote_server_uri" != "false" ]; then
                $SPLUNK_HOME/bin/splunk login -uri $remote_server_uri -auth $AUTH_STRING
                if [ "$?" != 0 ]; then
                    echo ""
                    echo " authentication failed"
                    echo ""
                    exit 1
                fi
            else
                $SPLUNK_HOME/bin/splunk login -auth $AUTH_STRING
                if [ "$?" != 0 ]; then
                    echo ""
                    echo " authentication failed"
                    echo ""
                    exit 1
                fi
            fi
        fi
        server_name=$(get_server_name)
        set_unix_app_info
        echo ""
        echo " authenticated to $server_name"
        echo ""
        _oldIFS=$IFS
        IFS="::"
        for key in $__QUEUE; do
            IFS=$_oldIFS
            execute_command "$key"
            IFS="::"
        done
        IFS=$_oldIFS
    fi
}

function queue_action
# creates queue of actions to be executed by execute_queue
{
    __QUEUE=$_QUEUE"::$ACTION $ACTION_TARGET "
}

### MAIN ###

. `dirname $0`/common.sh

remote_server_uri="false"
server_unix_app_installed="false"
server_has_unix_app_enabled="false"
remote_server_unix_app_installed="false"
remote_server_has_unix_app_enabled="false"

MONITOR_INPUTS="/Library/Logs ~/Library/Logs /var/log /var/adm /etc"

__QUEUE=
ACTION=
ACTION_TARGET=
AUTH_STRING=
REMOTE_URI=

if [ ! -n "$1" ]; then
    splunk_login
else
    while [ "$1" != "" ]; do
        case $1 in
            --auth ) shift; AUTH_STRING="$1"; shift;;
            --clone-all ) ACTION="clone"; queue_action; shift;;
            --disable-all ) ACTION="disable"; ACTION_TARGET="all"; queue_action; shift;;
            --disable-input ) ACTION="disable"; shift; ACTION_TARGET="$1"; queue_action; shift;;
            --enable-all ) ACTION="enable"; shift; ACTION_TARGET="$1"; ACTION_TARGET="all "$ACTION_TARGET;shift;ACTION_TARGET=$ACTION_TARGET" $1";shift;queue_action; shift;;
            --enable-input ) ACTION="enable"; shift; ACTION_TARGET="$1";shift; ACTION_TARGET="input "$ACTION_TARGET" $1";shift;ACTION_TARGET=$ACTION_TARGET" $1";shift;queue_action; shift;;
            --interval ) ACTION="interval"; shift; ACTION_TARGET="$1"; shift; ACTION_TARGET=$ACTION_TARGET" $1"; queue_action; shift;;
            --install-app ) ACTION="install"; shift; ACTION_TARGET="$1"; queue_action; shift;;
            --list-all ) ACTION="list"; ACTION_TARGET="all"; queue_action; shift;;
            --list-input ) ACTION="list"; shift; ACTION_TARGET="$1"; queue_action; shift;;
            --uri ) remote_server_uri="$1"; shift;;
            --usage | --help ) usage;;
            * ) usage;;
        esac
    done
    execute_queue
fi
