class DcmiReturnValueError(ValueError):
    """DCMI return value error"""
    pass


class ParamError(ValueError):
    """Parameter validation error"""
    pass


class TopoMissMatchError(ValueError):
    """Error: Failed to match super_pod_type with topology file"""
    pass


class GetIpError(ValueError):
    """Error: Failed to retrieve local IP address"""
    pass