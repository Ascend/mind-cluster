import enum


class ProductType(enum.Enum):
    """Super pod type code for all device forms. Standard card use its main board id."""
    SERVER_8P = 0
    POD_1D = 1
    POD_2D = 2
    SERVER_16P = 3
    STANDARD_1P = 104
    STANDARD_2P = 106
    STANDARD_4P = 108