# This custom fact pulls out all local groups from the /etc/group file
# and returns the collection as a comma-separated list.

Facter.add(:local_groups) do
  confine :kernel => 'Linux'
  setcode do
    groups = Array.new
    if File.exist?("/etc/group")
      File.open("/etc/group").each do |line|
        next if line.match(/^\s|^#|^$/)
        groups << line.split(':').first
      end
    end
    groups.join(',')
  end
end
