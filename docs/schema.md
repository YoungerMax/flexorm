## `text`

```json
{
    "type": "text",
    "primaryKey": false,
    "default": null,
    "length": null,
    "minLength": 1,
    "maxLength": null,
    "pattern": null,
}
```

- `default`: the string that should be
- `pattern`: a regex pattern that the string must be
- `length`: the length that the string MUST be (no more, no less)
- `minLength`: the minimum length the string can be
- `maxLength`: the maximum length the string can be, or null for unlimited

## `timestamp`

```json
{
    "type": "timestamp",
    "primaryKey": false,
    "default": null,
}
```

- `default`:
  - `null`: default value is null
  - `now`: default value is the current timestamp

## `integer`

```json
{
    "primaryKey": true,
    "type": "integer",
    "default": "autoincrement"
}
```

- `default`:
  - `autoincrement`: automatically increments
  - a number: that number is the default