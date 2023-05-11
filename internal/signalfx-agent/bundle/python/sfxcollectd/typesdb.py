"""
Logic dealing with types.db files
"""

from collections import namedtuple

DataSet = namedtuple("DataSet", "name sources")
DataSource = namedtuple("DataSource", "name type min max")

ACCEPTABLE_TYPES = ["GAUGE", "ABSOLUTE", "COUNTER", "DERIVE"]


def parse_types_db(content):
    """
    Returns a set of data source specifications parsed out of the given
    types.db file content.
    """
    data_sets = []
    for entry in content.splitlines():
        entry = entry.strip()

        # Tolerate blank/whitespace-only/comment lines
        if not entry or entry.startswith("#"):
            continue

        name, source_specs = entry.split()[0], entry.split()[1:]
        if not source_specs:
            raise ValueError("types.db entry has no data source specs: '%s'" % entry)

        dataset = DataSet(name=name.strip(), sources=[])
        for source in source_specs:
            parts = source.rstrip(",").split(":")
            if len(parts) != 4:
                raise ValueError("types.db data source '%s' is not a quadruple" % source)

            name, source_type, min_val, max_val = parts
            if source_type.upper() not in ACCEPTABLE_TYPES:
                raise ValueError("Bad type '%s' in data source spec '%s'" % (source_type, entry))

            dataset.sources.append(DataSource(name=name, type=source_type.upper(), min=min_val, max=max_val))

        data_sets.append(dataset)

    return data_sets
