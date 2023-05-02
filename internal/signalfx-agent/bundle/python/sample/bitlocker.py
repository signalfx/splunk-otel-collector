import logging
import pythoncom  # pylint: disable=import-error
import wmi  # pylint: disable=import-error

logger = logging.getLogger(__name__)


def run(_, output):
    # because run is called in a thread it is necessary to explicitly initialize
    # the COM libraries
    pythoncom.CoInitializeEx(pythoncom.COINIT_APARTMENTTHREADED)

    # if the bitlocker drive encryption feature is not installed then the attempt
    # to connect to the COM Object will raise an exception
    mve = None
    try:
        mve = wmi.WMI(moniker="//./root/cimv2/security/microsoftvolumeencryption")
    except Exception as e:  # pylint: disable=broad-except
        # BDE is not installed, report enabled = 0, locked = 0 for all drives
        logger.error("Error connecting to Bitlocker feature, assuming not installed: %s", e)
        mwmi = wmi.WMI()
        logical_disk_list = mwmi.Win32_LogicalDisk()
        for logical_disk in logical_disk_list:
            drive = logical_disk.DeviceID
            output.send_gauge("bitlocker_drive_encryption.enabled", 0, {"volume": drive})
            output.send_gauge("bitlocker_drive_encryption.locked", 0, {"volume": drive})
        return
    encryptable_volume_list = mve.Win32_EncryptableVolume()
    for encryptable_volume in encryptable_volume_list:
        drive = encryptable_volume.DriveLetter
        bde_enabled = 0
        bde_locked = 0
        if encryptable_volume.ProtectionStatus > 0:
            bde_enabled = 1
            bde_locked = 1 if (encryptable_volume.ProtectionStatus == 2) else 0
        output.send_gauge("bitlocker_drive_encryption.enabled", bde_enabled, {"volume": drive})
        output.send_gauge("bitlocker_drive_encryption.locked", bde_locked, {"volume": drive})
