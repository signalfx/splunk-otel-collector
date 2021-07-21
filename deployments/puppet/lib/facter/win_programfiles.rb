# Returns the PROGRAMFILES env var on windows

Facter.add(:win_programfiles) do
  confine :osfamily => :windows
  setcode do
    ENV['PROGRAMFILES']
  end
end
