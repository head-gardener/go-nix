package daemon

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/nix-community/go-nix/pkg/wire"
)

var ErrOptionalEmpty = errors.New("optional field is empty")

// readAck reads the daemon's acknowledgment uint64 and verifies it equals 1.
func readAck(dec *wire.Decoder) error {
	v, err := dec.ReadUint64()
	if err != nil {
		return err
	}

	if v != 1 {
		return &ProtocolError{
			Op:  "read ack",
			Err: fmt.Errorf("expected ack value 1, got %d", v),
		}
	}

	return nil
}

func (info *PathInfo) Convert(c wire.Codec) error {
	if err := c.String(&info.StorePath); err != nil {
		return err
	}

	if err := c.String(&info.Deriver); err != nil {
		return err
	}

	if err := c.String(&info.NarHash); err != nil {
		return err
	}

	if err := c.Strings(&info.References); err != nil {
		return err
	}

	if err := c.Uint64(&info.RegistrationTime); err != nil {
		return err
	}

	if err := c.Uint64(&info.NarSize); err != nil {
		return err
	}

	// Protocol >= 1.16: ultimate, sigs, ca.
	if c.Version() >= ProtoVersionPathInfoMeta {
		if err := c.Bool(&info.Ultimate); err != nil {
			return err
		}

		if err := c.Strings(&info.Sigs); err != nil {
			return err
		}

		if err := c.String(&info.CA); err != nil {
			return err
		}
	}

	return nil
}

// ReadPathInfo reads a full PathInfo from the wire (UnkeyedValidPathInfo format).
// The version parameter is the negotiated protocol version.
func ReadPathInfo(dec *wire.Decoder, version uint64) (*PathInfo, error) {
	info := PathInfo{}
	c := wire.NewDecoderNG(dec.Reader(), MaxStringSize, version)
	return &info, info.Convert(c)
}

// WritePathInfo writes a PathInfo in ValidPathInfo wire format.
// The version parameter is the negotiated protocol version.
func WritePathInfo(enc *wire.Encoder, info *PathInfo, version uint64) error {
	if info == nil {
		return ErrNilPathInfo
	}

	c := wire.NewEncoderNG(enc.Writer(), version)
	return info.Convert(c)
}

func (out *DerivationOutput) Convert(c wire.Codec) error {
	if err := c.String(&out.Path); err != nil {
		return err
	}

	if err := c.String(&out.HashAlgorithm); err != nil {
		return err
	}

	return c.String(&out.Hash)
}

func (drv *BasicDerivation) Convert(c wire.Codec) error {
	if err := wire.ConvertMap(c, &drv.Outputs); err != nil {
		return err
	}

	if err := c.Strings(&drv.Inputs); err != nil {
		return err
	}

	if err := c.String(&drv.Platform); err != nil {
		return err
	}

	if err := c.String(&drv.Builder); err != nil {
		return err
	}

	if err := c.Strings(&drv.Args); err != nil {
		return err
	}

	return c.StringMap(&drv.Env)
}

// WriteBasicDerivation writes a BasicDerivation to the wire.
// Outputs are written sorted by name.
// Environment variables are written sorted by key.
func WriteBasicDerivation(enc *wire.Encoder, drv *BasicDerivation) error {
	// nil check
	if drv == nil {
		return ErrNilDerivation
	}

	c := wire.NewEncoderNG(enc.Writer(), ProtoVersionFeatureExchange)

	return drv.Convert(c)
}

// readOptionalMicroseconds reads an optional<microseconds> from the wire.
// Wire format: tag(uint64: 0=none, 1=some) [+ value(uint64) if tag=1].
// Returns ErrOptionalEmpty if absent, or a pointer to the duration if present.
func readOptionalMicroseconds(dec *wire.Decoder) (*time.Duration, error) {
	tag, err := dec.ReadUint64()
	if err != nil {
		return nil, err
	}

	switch tag {
	case 0: // none
		return nil, ErrOptionalEmpty
	case optionalSome:
		us, err := dec.ReadUint64()
		if err != nil {
			return nil, err
		}

		d := time.Duration(us) * time.Microsecond //nolint:gosec // G115: microsecond values won't overflow int64

		return &d, nil
	default:
		return nil, &ProtocolError{
			Op:  "read optional microseconds",
			Err: fmt.Errorf("unexpected optional tag %d", tag),
		}
	}
}

// ReadBuildResult reads a BuildResult from the wire.
// The version parameter is the negotiated protocol version.
func ReadBuildResult(dec *wire.Decoder, version uint64) (*BuildResult, error) {
	status, err := dec.ReadUint64()
	if err != nil {
		return nil, &ProtocolError{Op: "read build result status", Err: err}
	}

	errorMsg, err := dec.ReadString()
	if err != nil {
		return nil, &ProtocolError{Op: "read build result errorMsg", Err: err}
	}

	result := &BuildResult{
		Status:   BuildStatus(status),
		ErrorMsg: errorMsg,
	}

	// Protocol >= 1.29: timing fields.
	if version >= ProtoVersionBuildTimes {
		result.TimesBuilt, err = dec.ReadUint64()
		if err != nil {
			return nil, &ProtocolError{Op: "read build result timesBuilt", Err: err}
		}

		result.IsNonDeterministic, err = dec.ReadBool()
		if err != nil {
			return nil, &ProtocolError{Op: "read build result isNonDeterministic", Err: err}
		}

		result.StartTime, err = dec.ReadUint64()
		if err != nil {
			return nil, &ProtocolError{Op: "read build result startTime", Err: err}
		}

		result.StopTime, err = dec.ReadUint64()
		if err != nil {
			return nil, &ProtocolError{Op: "read build result stopTime", Err: err}
		}
	}

	// Protocol >= 1.37: cpuUser and cpuSystem as optional<microseconds>.
	if version >= ProtoVersionCPUTimes {
		result.CpuUser, err = readOptionalMicroseconds(dec)
		if errors.Is(err, ErrOptionalEmpty) {
			// do nothing
		} else if err != nil {
			return nil, &ProtocolError{Op: "read build result cpuUser", Err: err}
		}

		result.CpuSystem, err = readOptionalMicroseconds(dec)
		if errors.Is(err, ErrOptionalEmpty) {
			// do nothing
		} else if err != nil {
			return nil, &ProtocolError{Op: "read build result cpuSystem", Err: err}
		}
	}

	// Protocol >= 1.28: builtOutputs map.
	if version >= ProtoVersionBuiltOutputs {
		nrOutputs, err := dec.ReadUint64()
		if err != nil {
			return nil, &ProtocolError{Op: "read build result builtOutputs count", Err: err}
		}

		builtOutputs := make(map[string]Realisation, nrOutputs)

		for range nrOutputs {
			name, err := dec.ReadString()
			if err != nil {
				return nil, &ProtocolError{Op: "read build result output name", Err: err}
			}

			realisationJSON, err := dec.ReadString()
			if err != nil {
				return nil, &ProtocolError{Op: "read build result realisation", Err: err}
			}

			var realisation Realisation
			if err := json.Unmarshal([]byte(realisationJSON), &realisation); err != nil {
				return nil, &ProtocolError{Op: "read build result realisation JSON", Err: err}
			}

			builtOutputs[name] = realisation
		}

		result.BuiltOutputs = builtOutputs
	}

	return result, nil
}
