if __name__ == '__main__':
    try:
        data = json.load(sys.stdin)
        _exploit_fn(data['host'], data['port'], data['flag_ids'])
    except SystemExit:
        raise
    except:
        traceback.print_exc(file=sys.stderr)
        sys.exit(1)
