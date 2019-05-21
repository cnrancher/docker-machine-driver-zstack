package instance

const AutoFdiskScript = `#/bin/bash
#fdisk ,formating and create the file system on /dev/vda or /dev/sda
DISK_ATTACH_POINT="/dev/vda"
MOUNT_PATH="/mnt/sda1"
DOCKER_DIR_PATH="/var/lib/docker"
#KUBELET_DIR_PATH="/var/lib/kubelet"
fdisk_fun()
{
fdisk -S 56 \$DISK_ATTACH_POINT << EOF
n
p
1


wq
EOF

sleep 5
mkfs.ext4 -i 8192 \${DISK_ATTACH_POINT}1
}

#config /etc/fstab and mount device
main()
{
  #if [ -b "/dev/sda" ]; then
  #  DISK_ATTACH_POINT="/dev/sda"
  #fi

  fdisk_fun
  flag=0
  if [ -d "/var/lib/docker" ];then
    flag=1
    /etc/init.d/docker stop
    rm -fr /var/lib/docker
  fi
  
  mkdir \${MOUNT_PATH}
  echo "\${DISK_ATTACH_POINT}1   \${MOUNT_PATH}  ext4    defaults        0 0" >>/etc/fstab
  mount -a
  mount --make-shared \${MOUNT_PATH}
  mount --make-shared /
  mkdir -p \${MOUNT_PATH}\${DOCKER_DIR_PATH}
  #mkdir -p \${MOUNT_PATH}\${KUBELET_DIR_PATH}

  ln -sf \${MOUNT_PATH}\${DOCKER_DIR_PATH} \${DOCKER_DIR_PATH}
  #ln -sf \${MOUNT_PATH}\${KUBELET_DIR_PATH} \${KUBELET_DIR_PATH}

  if [ \$flag==1 ]; then
    /etc/init.d/docker start
  fi
}

main
df -h

`
