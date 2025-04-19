package serverdebug

import (
	"go.uber.org/zap/zapcore"
)

type dualCore struct {
	core1 zapcore.Core
	core2 zapcore.Core
}

func (d *dualCore) Enabled(lvl zapcore.Level) bool {
	return d.core1.Enabled(lvl) || d.core2.Enabled(lvl)
}

func (d *dualCore) With(fields []zapcore.Field) zapcore.Core {
	return &dualCore{
		core1: d.core1.With(fields),
		core2: d.core2.With(fields),
	}
}

func (d *dualCore) Check(ent zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if d.core1.Enabled(ent.Level) {
		ce = d.core1.Check(ent, ce)
	}
	if d.core2.Enabled(ent.Level) {
		ce = d.core2.Check(ent, ce)
	}
	return ce
}

func (d *dualCore) Write(ent zapcore.Entry, fields []zapcore.Field) error {
	err1 := d.core1.Write(ent, fields)
	err2 := d.core2.Write(ent, fields)
	if err1 != nil {
		return err1
	}
	return err2
}

func (d *dualCore) Sync() error {
	err1 := d.core1.Sync()
	err2 := d.core2.Sync()
	if err1 != nil {
		return err1
	}
	return err2
}
