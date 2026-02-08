import abc


class DumpLogDirParser(abc.ABC):

    def __init__(self, root_dir: str, parse_dir: str):
        self.root_dir = root_dir
        self.parse_dir = parse_dir

    @abc.abstractmethod
    def parse(self) -> dict:
        pass
