# Returns the TEMP env var on windows

Facter.add(:win_temp) do
  confine :osfamily => :windows
  setcode do
    ENV['TEMP']
  end
end
