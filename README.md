# MDIG

MDIG is the simple multicast DNS lookup utility, similar to dig but
much simplified

## Usage

    Usage:
        mdig [@interface] [options] domain [q-type] [q-class]

    Options may be intermixed with other parameters.
    Use -- to terminate options list.

    The @interface specifies network interface (by name)
    If missed, all active interfaces are used

    Options are:
        -4         use IPv4 (the default, may be combined with -6)
        -6         use IPv6 (may be combined with -4)
        -d         enable debugging
        -v         enable verbose debugging
        -p period  MDNS query period, milliseconds (default is 250)
        -c count   MDNS query count, before exit (default is 10)
        -h         print help screen and exit

<!-- vim:ts=8:sw=4:et:tw=72:
-->


