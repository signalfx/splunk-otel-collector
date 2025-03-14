# Returns the PROGRAMDATA env var on windows

Facter.add(:win_programdata) do
  confine :kernel => 'windows'
  setcode do
    ENV['CSIDL_COMMON_APPDATA']
  end
end
