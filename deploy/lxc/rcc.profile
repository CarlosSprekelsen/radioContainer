# LXC Profile for Radio Control Container (RCC)
# Provides minimal security hardening for ARM32 LXE deployment targets.

name: rcc
description: RCC container profile (arm32)

config:
  environment.arch: armhf
  environment.ubuntu: jammy
  security.nesting: "false"
  security.privileged: "false"
  security.idmap.isolated: "true"
  limits.cpu: "2"
  limits.memory: 512MB
  limits.processes: "512"
  network.type: veth
  network.link: lxcbr0
  network.flags: up
  network.ipv4.address: auto
  network.ipv6.address: auto
  rootfs.size: 2GB

  raw.lxc: |
    lxc.cap.drop = sys_module sys_rawio sys_time sys_tty_config
    lxc.cap.drop = sys_admin mac_admin sys_boot
    lxc.apparmor.allow_incomplete = 0
    lxc.start.auto = 1
    lxc.start.delay = 5

environment:
  RCC_ADDR: 0.0.0.0:8080

# Mount points and volumes can be extended on-device.
devices:
  app-root:
    type: disk
    source: /opt/rcc
    path: /opt/rcc
    readonly: false
  logs:
    type: disk
    source: /var/log/rcc
    path: /var/log/rcc
    readonly: false
