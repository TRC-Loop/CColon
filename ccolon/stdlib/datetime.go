package stdlib

import (
	"fmt"
	"time"

	"github.com/TRC-Loop/ccolon/vm"
)

func NewDatetimeModule() *vm.ModuleValue {
	return &vm.ModuleValue{
		Name: "datetime",
		Methods: map[string]*vm.NativeFuncValue{
			"now": {
				Name: "datetime.now",
				Fn: func(args []vm.Value) (vm.Value, error) {
					return timeToDict(time.Now()), nil
				},
			},
			"utcnow": {
				Name: "datetime.utcnow",
				Fn: func(args []vm.Value) (vm.Value, error) {
					return timeToDict(time.Now().UTC()), nil
				},
			},
			"timestamp": {
				Name: "datetime.timestamp",
				Fn: func(args []vm.Value) (vm.Value, error) {
					return &vm.IntValue{Val: time.Now().Unix()}, nil
				},
			},
			"timestamp_ms": {
				Name: "datetime.timestamp_ms",
				Fn: func(args []vm.Value) (vm.Value, error) {
					return &vm.IntValue{Val: time.Now().UnixMilli()}, nil
				},
			},
			"from_timestamp": {
				Name: "datetime.from_timestamp",
				Fn: func(args []vm.Value) (vm.Value, error) {
					if len(args) != 1 {
						return nil, fmt.Errorf("datetime.from_timestamp() takes 1 argument, got %d", len(args))
					}
					ts, ok := args[0].(*vm.IntValue)
					if !ok {
						return nil, fmt.Errorf("datetime.from_timestamp() requires an int")
					}
					return timeToDict(time.Unix(ts.Val, 0)), nil
				},
			},
			"format": {
				Name: "datetime.format",
				Fn: func(args []vm.Value) (vm.Value, error) {
					if len(args) < 1 || len(args) > 2 {
						return nil, fmt.Errorf("datetime.format() takes 1-2 arguments, got %d", len(args))
					}
					d, ok := args[0].(*vm.DictValue)
					if !ok {
						return nil, fmt.Errorf("datetime.format() requires a datetime dict")
					}
					t, err := dictToTime(d)
					if err != nil {
						return nil, err
					}
					layout := "2006-01-02 15:04:05"
					if len(args) == 2 {
						s, ok := args[1].(*vm.StringValue)
						if !ok {
							return nil, fmt.Errorf("datetime.format() format must be a string")
						}
						layout = ccolonToGoLayout(s.Val)
					}
					return &vm.StringValue{Val: t.Format(layout)}, nil
				},
			},
			"parse": {
				Name: "datetime.parse",
				Fn: func(args []vm.Value) (vm.Value, error) {
					if len(args) < 1 || len(args) > 2 {
						return nil, fmt.Errorf("datetime.parse() takes 1-2 arguments, got %d", len(args))
					}
					s, ok := args[0].(*vm.StringValue)
					if !ok {
						return nil, fmt.Errorf("datetime.parse() requires a string")
					}
					layout := "2006-01-02 15:04:05"
					if len(args) == 2 {
						l, ok := args[1].(*vm.StringValue)
						if !ok {
							return nil, fmt.Errorf("datetime.parse() layout must be a string")
						}
						layout = ccolonToGoLayout(l.Val)
					}
					t, err := time.Parse(layout, s.Val)
					if err != nil {
						return nil, fmt.Errorf("datetime.parse() failed: %s", err)
					}
					return timeToDict(t), nil
				},
			},
			"timezone": {
				Name: "datetime.timezone",
				Fn: func(args []vm.Value) (vm.Value, error) {
					if len(args) != 2 {
						return nil, fmt.Errorf("datetime.timezone() takes 2 arguments, got %d", len(args))
					}
					d, ok := args[0].(*vm.DictValue)
					if !ok {
						return nil, fmt.Errorf("datetime.timezone() requires a datetime dict")
					}
					tz, ok := args[1].(*vm.StringValue)
					if !ok {
						return nil, fmt.Errorf("datetime.timezone() requires a timezone string")
					}
					t, err := dictToTime(d)
					if err != nil {
						return nil, err
					}
					loc, err := time.LoadLocation(tz.Val)
					if err != nil {
						return nil, fmt.Errorf("datetime.timezone() invalid timezone: %s", tz.Val)
					}
					return timeToDict(t.In(loc)), nil
				},
			},
			"diff": {
				Name: "datetime.diff",
				Fn: func(args []vm.Value) (vm.Value, error) {
					if len(args) != 2 {
						return nil, fmt.Errorf("datetime.diff() takes 2 arguments, got %d", len(args))
					}
					d1, ok := args[0].(*vm.DictValue)
					if !ok {
						return nil, fmt.Errorf("datetime.diff() requires datetime dicts")
					}
					d2, ok := args[1].(*vm.DictValue)
					if !ok {
						return nil, fmt.Errorf("datetime.diff() requires datetime dicts")
					}
					t1, err := dictToTime(d1)
					if err != nil {
						return nil, err
					}
					t2, err := dictToTime(d2)
					if err != nil {
						return nil, err
					}
					diff := t1.Sub(t2)
					return &vm.FloatValue{Val: diff.Seconds()}, nil
				},
			},
			"sleep": {
				Name: "datetime.sleep",
				Fn: func(args []vm.Value) (vm.Value, error) {
					if len(args) != 1 {
						return nil, fmt.Errorf("datetime.sleep() takes 1 argument, got %d", len(args))
					}
					switch v := args[0].(type) {
					case *vm.IntValue:
						time.Sleep(time.Duration(v.Val) * time.Millisecond)
					case *vm.FloatValue:
						time.Sleep(time.Duration(v.Val * float64(time.Millisecond)))
					default:
						return nil, fmt.Errorf("datetime.sleep() requires a number (milliseconds)")
					}
					return &vm.NilValue{}, nil
				},
			},
		},
	}
}

func timeToDict(t time.Time) *vm.DictValue {
	d := &vm.DictValue{
		Entries: map[string]vm.Value{
			"year":     &vm.IntValue{Val: int64(t.Year())},
			"month":    &vm.IntValue{Val: int64(t.Month())},
			"day":      &vm.IntValue{Val: int64(t.Day())},
			"hour":     &vm.IntValue{Val: int64(t.Hour())},
			"minute":   &vm.IntValue{Val: int64(t.Minute())},
			"second":   &vm.IntValue{Val: int64(t.Second())},
			"weekday":  &vm.StringValue{Val: t.Weekday().String()},
			"timezone": &vm.StringValue{Val: t.Location().String()},
		},
		Order: []string{"year", "month", "day", "hour", "minute", "second", "weekday", "timezone"},
	}
	return d
}

func dictToTime(d *vm.DictValue) (time.Time, error) {
	getInt := func(key string) (int, error) {
		v, ok := d.Entries[key]
		if !ok {
			return 0, fmt.Errorf("missing '%s' in datetime dict", key)
		}
		iv, ok := v.(*vm.IntValue)
		if !ok {
			return 0, fmt.Errorf("'%s' must be an int", key)
		}
		return int(iv.Val), nil
	}

	year, err := getInt("year")
	if err != nil {
		return time.Time{}, err
	}
	month, err := getInt("month")
	if err != nil {
		return time.Time{}, err
	}
	day, err := getInt("day")
	if err != nil {
		return time.Time{}, err
	}
	hour, err := getInt("hour")
	if err != nil {
		return time.Time{}, err
	}
	minute, err := getInt("minute")
	if err != nil {
		return time.Time{}, err
	}
	second, err := getInt("second")
	if err != nil {
		return time.Time{}, err
	}

	loc := time.Local
	if tz, ok := d.Entries["timezone"]; ok {
		if s, ok := tz.(*vm.StringValue); ok && s.Val != "" {
			if l, err := time.LoadLocation(s.Val); err == nil {
				loc = l
			}
		}
	}

	return time.Date(year, time.Month(month), day, hour, minute, second, 0, loc), nil
}

// ccolonToGoLayout converts common format tokens to Go layout.
// Supports: %Y, %m, %d, %H, %M, %S, %Z
func ccolonToGoLayout(fmt string) string {
	r := []byte(fmt)
	result := make([]byte, 0, len(r)*2)
	for i := 0; i < len(r); i++ {
		if r[i] == '%' && i+1 < len(r) {
			i++
			switch r[i] {
			case 'Y':
				result = append(result, "2006"...)
			case 'm':
				result = append(result, "01"...)
			case 'd':
				result = append(result, "02"...)
			case 'H':
				result = append(result, "15"...)
			case 'M':
				result = append(result, "04"...)
			case 'S':
				result = append(result, "05"...)
			case 'Z':
				result = append(result, "MST"...)
			case '%':
				result = append(result, '%')
			default:
				result = append(result, '%', r[i])
			}
		} else {
			result = append(result, r[i])
		}
	}
	return string(result)
}
