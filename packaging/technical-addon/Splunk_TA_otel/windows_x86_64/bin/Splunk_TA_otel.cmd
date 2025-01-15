@echo off
:: Set inputs.conf and other configuration variables (values, input names, possibly defaults, etc)
set "splunk_TA_otel_app_directory=%~dp0"
set "splunk_otel_process_name=otelcol_windows_amd64.exe"
set "splunk_otel_common_ps_name=Splunk_TA_otel_utils.ps1"

set "SPLUNK_OTEL_TA_HOME="
set "SPLUNK_OTEL_TA_PLATFORM_HOME="

set "SPLUNK_OTEL_FLAGS="
:: BEGIN AUTOGENERATED CODE
set "configd_name=configd"
set "configd_value="
set "discovery_name=discovery"
set "discovery_value="
set "discovery_properties_name=discovery_properties"
set "discovery_properties_value="
set "gomemlimit_name=gomemlimit"
set "gomemlimit_value="
set "splunk_api_url_name=splunk_api_url"
set "splunk_api_url_value="
set "splunk_bundle_dir_name=splunk_bundle_dir"
set "splunk_bundle_dir_value="
set "splunk_config_name=splunk_config"
set "splunk_config_value="
set "splunk_config_dir_name=splunk_config_dir"
set "splunk_config_dir_value="
set "splunk_collectd_dir_name=splunk_collectd_dir"
set "splunk_collectd_dir_value="
set "splunk_debug_config_server_name=splunk_debug_config_server"
set "splunk_debug_config_server_value="
set "splunk_config_yaml_name=splunk_config_yaml"
set "splunk_config_yaml_value="
set "splunk_gateway_url_name=splunk_gateway_url"
set "splunk_gateway_url_value="
set "splunk_hec_url_name=splunk_hec_url"
set "splunk_hec_url_value="
set "splunk_listen_interface_name=splunk_listen_interface"
set "splunk_listen_interface_value="
set "splunk_memory_limit_mib_name=splunk_memory_limit_mib"
set "splunk_memory_limit_mib_value="
set "splunk_memory_total_mib_name=splunk_memory_total_mib"
set "splunk_memory_total_mib_value="
set "splunk_otel_log_file_name=splunk_otel_log_file"
set "splunk_otel_log_file_value="
set "splunk_ingest_url_name=splunk_ingest_url"
set "splunk_ingest_url_value="
set "splunk_realm_name=splunk_realm"
set "splunk_realm_value="
set "splunk_access_token_file_name=splunk_access_token_file"
set "splunk_access_token_file_value="
:: END AUTOGENERATED CODE


echo on
echo "Starting Splunk TA Otel."
echo off

:: switch based on which arg was passed
if "%1%"=="" goto splunk_TA_otel_run_agent
if "%1%"=="--scheme" goto splunk_TA_otel_scheme
if "%1%"=="--validate-arguments" goto splunk_TA_otel_validate_arg
:: exit if no matching handler for argument
exit /b

:: main entry hooks

:splunk_TA_otel_scheme
setlocal
echo "display scheme called"
endlocal
exit /B

:splunk_TA_otel_validate_arg
setlocal
echo "validate args called"
endlocal
exit /B 0


:splunk_TA_otel_run_agent
setlocal
:: READING CONFIGURATION FROM STDIN
call :splunk_TA_otel_read_configs

call :splunk_TA_otel_log_msg "INFO" "Starting otel agent from %splunk_TA_otel_app_directory% with configuration file %splunk_config_value% and log file %splunk_otel_log_file_value%"
call :splunk_TA_otel_log_msg "INFO" "Logging TA notices to %splunk_TA_otel_log_file%"
:: By default, otel will register itself as a windows service. In context of the TA, we want splunk to manage our lifecycle, so turn this off.
set "NO_WINDOWS_SERVICE=1"

if "%SPLUNK_ACCESS_TOKEN%"=="" (
    call :splunk_TA_otel_log_msg "INFO" "Grabbing SPLUNK_ACCESS_TOKEN from file %splunk_access_token_file%"
    call :get_access_token
) else (
    call :splunk_TA_otel_log_msg "INFO" "Environment variable SPLUNK_ACCESS_TOKEN already set."
)


:: BEGIN AUTOGENERATED CODE
if "%gomemlimit_value%" == "" (
    call :splunk_TA_otel_log_msg "DEBUG" "Param %gomemlimit_name% not set"
) else (
    set "GOMEMLIMIT=%gomemlimit_value%"
)
if "%splunk_api_url_value%" == "" (
    call :splunk_TA_otel_log_msg "DEBUG" "Param %splunk_api_url_name% not set"
) else (
    set "SPLUNK_API_URL=%splunk_api_url_value%"
)
if "%splunk_bundle_dir_value%" == "" (
    call :splunk_TA_otel_log_msg "DEBUG" "Param %splunk_bundle_dir_name% not set"
) else (
    set "SPLUNK_BUNDLE_DIR=%splunk_bundle_dir_value%"
)
if "%splunk_config_value%" == "" (
    call :splunk_TA_otel_log_msg "DEBUG" "Param %splunk_config_name% not set"
) else (
    set "SPLUNK_OTEL_FLAGS=%SPLUNK_OTEL_FLAGS% --config=%splunk_config_value%"
    set "SPLUNK_CONFIG=%splunk_config_value%"
)
if "%splunk_config_dir_value%" == "" (
    call :splunk_TA_otel_log_msg "DEBUG" "Param %splunk_config_dir_name% not set"
) else (
    set "SPLUNK_OTEL_FLAGS=%SPLUNK_OTEL_FLAGS% --config-dir=%splunk_config_dir_value%"
    set "SPLUNK_CONFIG_DIR=%splunk_config_dir_value%"
)
if "%splunk_collectd_dir_value%" == "" (
    call :splunk_TA_otel_log_msg "DEBUG" "Param %splunk_collectd_dir_name% not set"
) else (
    set "SPLUNK_COLLECTD_DIR=%splunk_collectd_dir_value%"
)
if "%splunk_debug_config_server_value%" == "" (
    call :splunk_TA_otel_log_msg "DEBUG" "Param %splunk_debug_config_server_name% not set"
) else (
    set "SPLUNK_DEBUG_CONFIG_SERVER=%splunk_debug_config_server_value%"
)
if "%splunk_config_yaml_value%" == "" (
    call :splunk_TA_otel_log_msg "DEBUG" "Param %splunk_config_yaml_name% not set"
) else (
    set "SPLUNK_CONFIG_YAML=%splunk_config_yaml_value%"
)
if "%splunk_gateway_url_value%" == "" (
    call :splunk_TA_otel_log_msg "DEBUG" "Param %splunk_gateway_url_name% not set"
) else (
    set "SPLUNK_GATEWAY_URL=%splunk_gateway_url_value%"
)
if "%splunk_hec_url_value%" == "" (
    call :splunk_TA_otel_log_msg "DEBUG" "Param %splunk_hec_url_name% not set"
) else (
    set "SPLUNK_HEC_URL=%splunk_hec_url_value%"
)
if "%splunk_listen_interface_value%" == "" (
    call :splunk_TA_otel_log_msg "DEBUG" "Param %splunk_listen_interface_name% not set"
) else (
    set "SPLUNK_LISTEN_INTERFACE=%splunk_listen_interface_value%"
)
if "%splunk_memory_limit_mib_value%" == "" (
    call :splunk_TA_otel_log_msg "DEBUG" "Param %splunk_memory_limit_mib_name% not set"
) else (
    set "SPLUNK_MEMORY_LIMIT_MIB=%splunk_memory_limit_mib_value%"
)
if "%splunk_memory_total_mib_value%" == "" (
    call :splunk_TA_otel_log_msg "DEBUG" "Param %splunk_memory_total_mib_name% not set"
) else (
    set "SPLUNK_MEMORY_TOTAL_MIB=%splunk_memory_total_mib_value%"
)
if "%splunk_otel_log_file_value%" == "" (
    call :splunk_TA_otel_log_msg "DEBUG" "Param %splunk_otel_log_file_name% not set"
) else (
    set "SPLUNK_OTEL_LOG_FILE=%splunk_otel_log_file_value%"
)
if "%splunk_ingest_url_value%" == "" (
    call :splunk_TA_otel_log_msg "DEBUG" "Param %splunk_ingest_url_name% not set"
) else (
    set "SPLUNK_INGEST_URL=%splunk_ingest_url_value%"
)
if "%splunk_realm_value%" == "" (
    call :splunk_TA_otel_log_msg "DEBUG" "Param %splunk_realm_name% not set"
) else (
    set "SPLUNK_REALM=%splunk_realm_value%"
)
if "%splunk_access_token_file_value%" == "" (
    call :splunk_TA_otel_log_msg "DEBUG" "Param %splunk_access_token_file_name% not set"
) else (
    set "SPLUNK_ACCESS_TOKEN_FILE=%splunk_access_token_file_value%"
)
if "%configd_value%" == "true" (
    set "SPLUNK_OTEL_FLAGS=%SPLUNK_OTEL_FLAGS% --configd"
) else (
    call :splunk_TA_otel_log_msg "DEBUG" "Optional flag %configd_name% not set"
)
if "%discovery_value%" == "true" (
    set "SPLUNK_OTEL_FLAGS=%SPLUNK_OTEL_FLAGS% --discovery"
) else (
    call :splunk_TA_otel_log_msg "DEBUG" "Optional flag %discovery_name% not set"
)
if "%discovery_properties_value%" != "" (
    set "SPLUNK_OTEL_FLAGS=%SPLUNK_OTEL_FLAGS% --discovery-properties=%discovery_properties_value%"
) else (
    call :splunk_TA_otel_log_msg "DEBUG" "Optional flag %discovery_properties_name% not set"
)
:: END AUTOGENERATED CODE

:: Extract agent bundle
call :extract_bundle

set "command_line=%splunk_TA_otel_app_directory%%splunk_otel_process_name%"
start /B "" "%command_line%" "%SPLUNK_OTEL_FLAGS%" > "%splunk_otel_log_file_value%" 2>&1
set "splunk_otel_common_ps=%splunk_TA_otel_app_directory%%splunk_otel_common_ps_name%"
start /B "" /I /WAIT powershell "& '%splunk_otel_common_ps%' '%splunk_otel_process_name%' '%splunk_TA_otel_log_file%'"

call :splunk_TA_otel_log_msg "INFO" "Otel agent stopped"
endlocal
exit /B 0

:: Helper functions

:splunk_TA_otel_log_msg
setlocal
echo on
set "log_type=%~1"
set "log_msg=%~2"

for /f "delims=" %%a in ('powershell -noninteractive -noprofile -command "get-date -format 'MM-dd-yyyy HH:mm K'"') do (
    set "log_date=%%a"
)

if not "%log_type%%SPLUNK_OTEL_TA_DEBUG%" == "DEBUG" (
    echo "%log_date%" "%log_type%" "%log_msg%" >> "%splunk_TA_otel_log_file%"
)

endlocal
exit /B 0

:splunk_TA_otel_read_configs
echo "INFO grabbing config from stdin..."



for /F "tokens=1,2 delims==" %%I in ('powershell -noninteractive -noprofile -command "$input | Select-String -Pattern '.*?(%configd_name%|%discovery_name%|%discovery_properties_name%|%gomemlimit_name%|%splunk_access_token_file_name%|%splunk_api_url_name%|%splunk_ballast_size_mib_name%|%splunk_bundle_dir_name%|%splunk_collectd_dir_name%|%splunk_config_name%|%splunk_config_dir_name%|%splunk_config_yaml_name%|%splunk_debug_config_server_name%|%splunk_gateway_url_name%|%splunk_hec_url_name%|%splunk_ingest_url_name%|%splunk_listen_interface_name%|%splunk_memory_limit_mib_name%|%splunk_memory_total_mib_name%|%splunk_otel_log_file_name%|%splunk_realm_name%).*?>(.*?)<' | ForEach-Object { $_.Matches.Groups[1].Value + '=' + $_.Matches.Groups[2].Value }"') do (
    if "%%I"=="%configd_name%" set "configd_value=%%J"
    if "%%I"=="%discovery_name%" set "discovery_value=%%J"
    if "%%I"=="%discovery_properties_name%" set "discovery_properties_value=%%J"
    if "%%I"=="%gomemlimit_name%" set "gomemlimit_value=%%J"
    if "%%I"=="%splunk_api_url_name%" set "splunk_api_url_value=%%J"
    if "%%I"=="%splunk_bundle_dir_name%" set "splunk_bundle_dir_value=%%J"
    if "%%I"=="%splunk_config_name%" set "splunk_config_value=%%J"
    if "%%I"=="%splunk_config_dir_name%" set "splunk_config_dir_value=%%J"
    if "%%I"=="%splunk_collectd_dir_name%" set "splunk_collectd_dir_value=%%J"
    if "%%I"=="%splunk_config_yaml_name%" set "splunk_config_yaml_value=%%J"
    if "%%I"=="%splunk_debug_config_server_name%" set "splunk_debug_config_server_value=%%J"
    if "%%I"=="%splunk_gateway_url_name%" set "splunk_gateway_url_value=%%J"
    if "%%I"=="%splunk_hec_url_name%" set "splunk_hec_url_value=%%J"
    if "%%I"=="%splunk_listen_interface_name%" set "splunk_listen_interface_value=%%J"
    if "%%I"=="%splunk_memory_limit_mib_name%" set "splunk_memory_limit_mib_value=%%J"
    if "%%I"=="%splunk_memory_total_mib_name%" set "splunk_memory_total_mib_value=%%J"
    if "%%I"=="%splunk_otel_log_file_name%" set "splunk_otel_log_file_value=%%J"
    if "%%I"=="%splunk_ingest_url_name%" set "splunk_ingest_url_value=%%J"
    if "%%I"=="%splunk_realm_name%" set "splunk_realm_value=%%J"
    if "%%I"=="%splunk_access_token_file_name%" set "splunk_access_token_file_value=%%J"
)
echo "INFO done grabbing configs..."

if "%SPLUNK_OTEL_TA_PLATFORM_HOME%"=="" (
    for /f "delims=" %%i in ('powershell  -noninteractive -noprofile -command "(get-item '%splunk_TA_otel_app_directory%' ).parent.FullName"') do (
        set "SPLUNK_OTEL_TA_PLATFORM_HOME=%%i"
    )
)
if "%SPLUNK_OTEL_TA_HOME%"=="" (
    for  /f "delims=" %%i in ('powershell  -noninteractive -noprofile -command "(get-item '%SPLUNK_OTEL_TA_PLATFORM_HOME%' ).parent.FullName"') do (
        set "SPLUNK_OTEL_TA_HOME=%%i"
    )
)
if "%SPLUNK_HOME%"=="" (
    :: Parent 3 times... $SPLUNK_HOME/etc/deployment-apps/Splunk_TA_otel
    for  /f "delims=" %%i in ('powershell  -noninteractive -noprofile -command "(get-item '%SPLUNK_OTEL_TA_HOME%' ).parent.parent.parent.FullName"') do (
        set "SPLUNK_HOME=%%i"
    )
) else (
    echo "SPLUNK_HOME already set, reusing value: %SPLUNK_HOME%"
)

set "splunk_TA_otel_log_file=%SPLUNK_HOME%\var\log\splunk\Splunk_TA_otel.log"
echo "otel ta log file: %splunk_TA_otel_log_file%"
call :splunk_TA_otel_log_msg "INFO" "splunk_TA_otel_app_directory: %splunk_TA_otel_app_directory%"

:: BEGIN AUTOGENERATED CODE
:: expand params in splunk_access_token_file_value
for /f "delims=" %%i in ('powershell -noninteractive -noprofile -command "'%splunk_access_token_file_value%' -replace '\$SPLUNK_OTEL_TA_PLATFORM_HOME', '%SPLUNK_OTEL_TA_PLATFORM_HOME%'"') do (
    set "splunk_access_token_file_value=%%i"
)
for /f "delims=" %%i in ('powershell -noninteractive -noprofile -command "'%splunk_access_token_file_value%' -replace '\$SPLUNK_OTEL_TA_HOME', '%SPLUNK_OTEL_TA_HOME%'"') do (
    set "splunk_access_token_file_value=%%i"
)
for /f "delims=" %%i in ('powershell -noninteractive -noprofile -command "'%splunk_access_token_file_value%' -replace '\$SPLUNK_HOME', '%SPLUNK_HOME%'"') do (
    set "splunk_access_token_file_value=%%i"
)


:: expand params in splunk_bundle_dir_value
for /f "delims=" %%i in ('powershell -noninteractive -noprofile -command "'%splunk_bundle_dir_value%' -replace '\$SPLUNK_OTEL_TA_PLATFORM_HOME', '%SPLUNK_OTEL_TA_PLATFORM_HOME%'"') do (
    set "splunk_bundle_dir_value=%%i"
)
for /f "delims=" %%i in ('powershell -noninteractive -noprofile -command "'%splunk_bundle_dir_value%' -replace '\$SPLUNK_OTEL_TA_HOME', '%SPLUNK_OTEL_TA_HOME%'"') do (
    set "splunk_bundle_dir_value=%%i"
)
for /f "delims=" %%i in ('powershell -noninteractive -noprofile -command "'%splunk_bundle_dir_value%' -replace '\$SPLUNK_HOME', '%SPLUNK_HOME%'"') do (
    set "splunk_bundle_dir_value=%%i"
)


:: expand params in splunk_config_value
for /f "delims=" %%i in ('powershell -noninteractive -noprofile -command "'%splunk_config_value%' -replace '\$SPLUNK_OTEL_TA_PLATFORM_HOME', '%SPLUNK_OTEL_TA_PLATFORM_HOME%'"') do (
    set "splunk_config_value=%%i"
)
for /f "delims=" %%i in ('powershell -noninteractive -noprofile -command "'%splunk_config_value%' -replace '\$SPLUNK_OTEL_TA_HOME', '%SPLUNK_OTEL_TA_HOME%'"') do (
    set "splunk_config_value=%%i"
)
for /f "delims=" %%i in ('powershell -noninteractive -noprofile -command "'%splunk_config_value%' -replace '\$SPLUNK_HOME', '%SPLUNK_HOME%'"') do (
    set "splunk_config_value=%%i"
)


:: expand params in splunk_config_dir_value
for /f "delims=" %%i in ('powershell -noninteractive -noprofile -command "'%splunk_config_dir_value%' -replace '\$SPLUNK_OTEL_TA_PLATFORM_HOME', '%SPLUNK_OTEL_TA_PLATFORM_HOME%'"') do (
    set "splunk_config_dir_value=%%i"
)
for /f "delims=" %%i in ('powershell -noninteractive -noprofile -command "'%splunk_config_dir_value%' -replace '\$SPLUNK_OTEL_TA_HOME', '%SPLUNK_OTEL_TA_HOME%'"') do (
    set "splunk_config_dir_value=%%i"
)
for /f "delims=" %%i in ('powershell -noninteractive -noprofile -command "'%splunk_config_dir_value%' -replace '\$SPLUNK_HOME', '%SPLUNK_HOME%'"') do (
    set "splunk_config_dir_value=%%i"
)


:: expand params in splunk_collectd_dir_value
for /f "delims=" %%i in ('powershell -noninteractive -noprofile -command "'%splunk_collectd_dir_value%' -replace '\$SPLUNK_OTEL_TA_PLATFORM_HOME', '%SPLUNK_OTEL_TA_PLATFORM_HOME%'"') do (
    set "splunk_collectd_dir_value=%%i"
)
for /f "delims=" %%i in ('powershell -noninteractive -noprofile -command "'%splunk_collectd_dir_value%' -replace '\$SPLUNK_OTEL_TA_HOME', '%SPLUNK_OTEL_TA_HOME%'"') do (
    set "splunk_collectd_dir_value=%%i"
)
for /f "delims=" %%i in ('powershell -noninteractive -noprofile -command "'%splunk_collectd_dir_value%' -replace '\$SPLUNK_HOME', '%SPLUNK_HOME%'"') do (
    set "splunk_collectd_dir_value=%%i"
)


:: expand params in splunk_config_yaml_value
for /f "delims=" %%i in ('powershell -noninteractive -noprofile -command "'%splunk_config_yaml_value%' -replace '\$SPLUNK_OTEL_TA_PLATFORM_HOME', '%SPLUNK_OTEL_TA_PLATFORM_HOME%'"') do (
    set "splunk_config_yaml_value=%%i"
)
for /f "delims=" %%i in ('powershell -noninteractive -noprofile -command "'%splunk_config_yaml_value%' -replace '\$SPLUNK_OTEL_TA_HOME', '%SPLUNK_OTEL_TA_HOME%'"') do (
    set "splunk_config_yaml_value=%%i"
)
for /f "delims=" %%i in ('powershell -noninteractive -noprofile -command "'%splunk_config_yaml_value%' -replace '\$SPLUNK_HOME', '%SPLUNK_HOME%'"') do (
    set "splunk_config_yaml_value=%%i"
)

:: expand params in splunk_otel_log_file_value
for /f "delims=" %%i in ('powershell -noninteractive -noprofile -command "'%splunk_otel_log_file_value%' -replace '\$SPLUNK_OTEL_TA_PLATFORM_HOME', '%SPLUNK_OTEL_TA_PLATFORM_HOME%'"') do (
    set "splunk_otel_log_file_value=%%i"
)
for /f "delims=" %%i in ('powershell -noninteractive -noprofile -command "'%splunk_otel_log_file_value%' -replace '\$SPLUNK_OTEL_TA_HOME', '%SPLUNK_OTEL_TA_HOME%'"') do (
    set "splunk_otel_log_file_value=%%i"
)
for /f "delims=" %%i in ('powershell -noninteractive -noprofile -command "'%splunk_otel_log_file_value%' -replace '\$SPLUNK_HOME', '%SPLUNK_HOME%'"') do (
    set "splunk_otel_log_file_value=%%i"
)

:: expand params in discovery_properties_value
for /f "delims=" %%i in ('powershell -noninteractive -noprofile -command "'%discovery_properties_value%' -replace '\$SPLUNK_OTEL_TA_PLATFORM_HOME', '%SPLUNK_OTEL_TA_PLATFORM_HOME%'"') do (
    set "discovery_properties_value=%%i"
)
for /f "delims=" %%i in ('powershell -noninteractive -noprofile -command "'%discovery_properties_value%' -replace '\$SPLUNK_OTEL_TA_HOME', '%SPLUNK_OTEL_TA_HOME%'"') do (
    set "discovery_properties_value=%%i"
)
for /f "delims=" %%i in ('powershell -noninteractive -noprofile -command "'%discovery_properties_value%' -replace '\$SPLUNK_HOME', '%SPLUNK_HOME%'"') do (
    set "discovery_properties_value=%%i"
)
:: END AUTOGENERATED CODE

:get_access_token
set /p SPLUNK_ACCESS_TOKEN=<"%splunk_access_token_file_value%"
exit /B 0

:extract_bundle
call :splunk_TA_otel_log_msg "INFO" "Extract agent bundle from '%splunk_TA_otel_app_directory%\agent-bundle_windows_amd64.zip' to %splunk_bundle_dir%"
for /f "delims=" %%i in ('powershell -noninteractive -noprofile -command "if (-not (Test-Path '%splunk_bundle_dir%')) {Expand-Archive -LiteralPath '%splunk_TA_otel_app_directory%\agent-bundle_windows_amd64.zip' -Destination '%splunk_bundle_dir%\..\' -Force }"') do (
    call :splunk_TA_otel_log_msg "DEBUG" "result from extract: %%i"
)
exit /B 0
