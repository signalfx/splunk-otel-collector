# This custom fact checks for the installed collector path on windows.
# Returns empty string if the key or the path does not exist.

Facter.add(:win_collector_path) do
  confine :osfamily => :windows
  setcode do
    begin
      value = ''
      Win32::Registry::HKEY_LOCAL_MACHINE.open('SYSTEM\CurrentControlSet\Services\splunk-otel-collector') do |regkey|
        value = regkey['ExePath']
      end
      if value and !File.exist?(value)
        value = ''
      end
      value
    rescue
      ''
    end
  end
end
