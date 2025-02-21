# Returns the PROGRAMFILES env var on windows

Facter.add(:win_programfiles) do
  confine :kernel => 'windows'
  setcode do
    ENV['CSIDL_PROGRAM_FILES']
  end
end
