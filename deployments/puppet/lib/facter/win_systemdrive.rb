# Returns the SystemDrive env var on windows

Facter.add(:win_systemdrive) do
  confine :kernel => 'windows'
  setcode do
    ENV['SystemDrive']
  end
end
