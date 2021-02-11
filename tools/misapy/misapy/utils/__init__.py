def struct_x_included_in_y(x, y, except_fields=[]):
    if type(x) != type(y):
        return False

    if type(x) == dict:
        if all(struct_x_included_in_y(x[key], y[key]) for key in x if key not in except_fields):
            return True
        else:
            return False
    # TODO implement comparison for lists
    # (with optional enforcement of order)
    else:
        return True if x==y else False