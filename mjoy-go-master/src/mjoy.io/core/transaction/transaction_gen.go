package transaction

// NOTE: THIS FILE WAS PRODUCED BY THE
// MSGP CODE GENERATION TOOL (github.com/tinylib/msgp)
// DO NOT EDIT

import (
	"github.com/tinylib/msgp/msgp"
	"mjoy.io/common/types"
)

// DecodeMsg implements msgp.Decodable
func (z *Action) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "Address":
			if dc.IsNil() {
				err = dc.ReadNil()
				if err != nil {
					return
				}
				z.Address = nil
			} else {
				if z.Address == nil {
					z.Address = new(types.Address)
				}
				err = z.Address.DecodeMsg(dc)
				if err != nil {
					return
				}
			}
		case "Params":
			z.Params, err = dc.ReadBytes(z.Params)
			if err != nil {
				return
			}
		default:
			err = dc.Skip()
			if err != nil {
				return
			}
		}
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z *Action) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 2
	// write "Address"
	err = en.Append(0x82, 0xa7, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73)
	if err != nil {
		return
	}
	if z.Address == nil {
		err = en.WriteNil()
		if err != nil {
			return
		}
	} else {
		err = z.Address.EncodeMsg(en)
		if err != nil {
			return
		}
	}
	// write "Params"
	err = en.Append(0xa6, 0x50, 0x61, 0x72, 0x61, 0x6d, 0x73)
	if err != nil {
		return
	}
	err = en.WriteBytes(z.Params)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *Action) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 2
	// string "Address"
	o = append(o, 0x82, 0xa7, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73)
	if z.Address == nil {
		o = msgp.AppendNil(o)
	} else {
		o, err = z.Address.MarshalMsg(o)
		if err != nil {
			return
		}
	}
	// string "Params"
	o = append(o, 0xa6, 0x50, 0x61, 0x72, 0x61, 0x6d, 0x73)
	o = msgp.AppendBytes(o, z.Params)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *Action) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "Address":
			if msgp.IsNil(bts) {
				bts, err = msgp.ReadNilBytes(bts)
				if err != nil {
					return
				}
				z.Address = nil
			} else {
				if z.Address == nil {
					z.Address = new(types.Address)
				}
				bts, err = z.Address.UnmarshalMsg(bts)
				if err != nil {
					return
				}
			}
		case "Params":
			z.Params, bts, err = msgp.ReadBytesBytes(bts, z.Params)
			if err != nil {
				return
			}
		default:
			bts, err = msgp.Skip(bts)
			if err != nil {
				return
			}
		}
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z *Action) Msgsize() (s int) {
	s = 1 + 8
	if z.Address == nil {
		s += msgp.NilSize
	} else {
		s += z.Address.Msgsize()
	}
	s += 7 + msgp.BytesPrefixSize + len(z.Params)
	return
}

// DecodeMsg implements msgp.Decodable
func (z *ActionSlice) DecodeMsg(dc *msgp.Reader) (err error) {
	var zb0002 uint32
	zb0002, err = dc.ReadArrayHeader()
	if err != nil {
		return
	}
	if cap((*z)) >= int(zb0002) {
		(*z) = (*z)[:zb0002]
	} else {
		(*z) = make(ActionSlice, zb0002)
	}
	for zb0001 := range *z {
		var field []byte
		_ = field
		var zb0003 uint32
		zb0003, err = dc.ReadMapHeader()
		if err != nil {
			return
		}
		for zb0003 > 0 {
			zb0003--
			field, err = dc.ReadMapKeyPtr()
			if err != nil {
				return
			}
			switch msgp.UnsafeString(field) {
			case "Address":
				if dc.IsNil() {
					err = dc.ReadNil()
					if err != nil {
						return
					}
					(*z)[zb0001].Address = nil
				} else {
					if (*z)[zb0001].Address == nil {
						(*z)[zb0001].Address = new(types.Address)
					}
					err = (*z)[zb0001].Address.DecodeMsg(dc)
					if err != nil {
						return
					}
				}
			case "Params":
				(*z)[zb0001].Params, err = dc.ReadBytes((*z)[zb0001].Params)
				if err != nil {
					return
				}
			default:
				err = dc.Skip()
				if err != nil {
					return
				}
			}
		}
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z ActionSlice) EncodeMsg(en *msgp.Writer) (err error) {
	err = en.WriteArrayHeader(uint32(len(z)))
	if err != nil {
		return
	}
	for zb0004 := range z {
		// map header, size 2
		// write "Address"
		err = en.Append(0x82, 0xa7, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73)
		if err != nil {
			return
		}
		if z[zb0004].Address == nil {
			err = en.WriteNil()
			if err != nil {
				return
			}
		} else {
			err = z[zb0004].Address.EncodeMsg(en)
			if err != nil {
				return
			}
		}
		// write "Params"
		err = en.Append(0xa6, 0x50, 0x61, 0x72, 0x61, 0x6d, 0x73)
		if err != nil {
			return
		}
		err = en.WriteBytes(z[zb0004].Params)
		if err != nil {
			return
		}
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z ActionSlice) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	o = msgp.AppendArrayHeader(o, uint32(len(z)))
	for zb0004 := range z {
		// map header, size 2
		// string "Address"
		o = append(o, 0x82, 0xa7, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73)
		if z[zb0004].Address == nil {
			o = msgp.AppendNil(o)
		} else {
			o, err = z[zb0004].Address.MarshalMsg(o)
			if err != nil {
				return
			}
		}
		// string "Params"
		o = append(o, 0xa6, 0x50, 0x61, 0x72, 0x61, 0x6d, 0x73)
		o = msgp.AppendBytes(o, z[zb0004].Params)
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *ActionSlice) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var zb0002 uint32
	zb0002, bts, err = msgp.ReadArrayHeaderBytes(bts)
	if err != nil {
		return
	}
	if cap((*z)) >= int(zb0002) {
		(*z) = (*z)[:zb0002]
	} else {
		(*z) = make(ActionSlice, zb0002)
	}
	for zb0001 := range *z {
		var field []byte
		_ = field
		var zb0003 uint32
		zb0003, bts, err = msgp.ReadMapHeaderBytes(bts)
		if err != nil {
			return
		}
		for zb0003 > 0 {
			zb0003--
			field, bts, err = msgp.ReadMapKeyZC(bts)
			if err != nil {
				return
			}
			switch msgp.UnsafeString(field) {
			case "Address":
				if msgp.IsNil(bts) {
					bts, err = msgp.ReadNilBytes(bts)
					if err != nil {
						return
					}
					(*z)[zb0001].Address = nil
				} else {
					if (*z)[zb0001].Address == nil {
						(*z)[zb0001].Address = new(types.Address)
					}
					bts, err = (*z)[zb0001].Address.UnmarshalMsg(bts)
					if err != nil {
						return
					}
				}
			case "Params":
				(*z)[zb0001].Params, bts, err = msgp.ReadBytesBytes(bts, (*z)[zb0001].Params)
				if err != nil {
					return
				}
			default:
				bts, err = msgp.Skip(bts)
				if err != nil {
					return
				}
			}
		}
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z ActionSlice) Msgsize() (s int) {
	s = msgp.ArrayHeaderSize
	for zb0004 := range z {
		s += 1 + 8
		if z[zb0004].Address == nil {
			s += msgp.NilSize
		} else {
			s += z[zb0004].Address.Msgsize()
		}
		s += 7 + msgp.BytesPrefixSize + len(z[zb0004].Params)
	}
	return
}

// DecodeMsg implements msgp.Decodable
func (z *Transaction) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "Data":
			err = z.Data.DecodeMsg(dc)
			if err != nil {
				return
			}
		default:
			err = dc.Skip()
			if err != nil {
				return
			}
		}
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z *Transaction) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 1
	// write "Data"
	err = en.Append(0x81, 0xa4, 0x44, 0x61, 0x74, 0x61)
	if err != nil {
		return
	}
	err = z.Data.EncodeMsg(en)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *Transaction) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 1
	// string "Data"
	o = append(o, 0x81, 0xa4, 0x44, 0x61, 0x74, 0x61)
	o, err = z.Data.MarshalMsg(o)
	if err != nil {
		return
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *Transaction) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "Data":
			bts, err = z.Data.UnmarshalMsg(bts)
			if err != nil {
				return
			}
		default:
			bts, err = msgp.Skip(bts)
			if err != nil {
				return
			}
		}
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z *Transaction) Msgsize() (s int) {
	s = 1 + 5 + z.Data.Msgsize()
	return
}

// DecodeMsg implements msgp.Decodable
func (z *TransactionForProducing) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		default:
			err = dc.Skip()
			if err != nil {
				return
			}
		}
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z TransactionForProducing) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 0
	err = en.Append(0x80)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z TransactionForProducing) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 0
	o = append(o, 0x80)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *TransactionForProducing) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		default:
			bts, err = msgp.Skip(bts)
			if err != nil {
				return
			}
		}
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z TransactionForProducing) Msgsize() (s int) {
	s = 1
	return
}

// DecodeMsg implements msgp.Decodable
func (z *Transactions) DecodeMsg(dc *msgp.Reader) (err error) {
	var zb0002 uint32
	zb0002, err = dc.ReadArrayHeader()
	if err != nil {
		return
	}
	if cap((*z)) >= int(zb0002) {
		(*z) = (*z)[:zb0002]
	} else {
		(*z) = make(Transactions, zb0002)
	}
	for zb0001 := range *z {
		if dc.IsNil() {
			err = dc.ReadNil()
			if err != nil {
				return
			}
			(*z)[zb0001] = nil
		} else {
			if (*z)[zb0001] == nil {
				(*z)[zb0001] = new(Transaction)
			}
			var field []byte
			_ = field
			var zb0003 uint32
			zb0003, err = dc.ReadMapHeader()
			if err != nil {
				return
			}
			for zb0003 > 0 {
				zb0003--
				field, err = dc.ReadMapKeyPtr()
				if err != nil {
					return
				}
				switch msgp.UnsafeString(field) {
				case "Data":
					err = (*z)[zb0001].Data.DecodeMsg(dc)
					if err != nil {
						return
					}
				default:
					err = dc.Skip()
					if err != nil {
						return
					}
				}
			}
		}
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z Transactions) EncodeMsg(en *msgp.Writer) (err error) {
	err = en.WriteArrayHeader(uint32(len(z)))
	if err != nil {
		return
	}
	for zb0004 := range z {
		if z[zb0004] == nil {
			err = en.WriteNil()
			if err != nil {
				return
			}
		} else {
			// map header, size 1
			// write "Data"
			err = en.Append(0x81, 0xa4, 0x44, 0x61, 0x74, 0x61)
			if err != nil {
				return
			}
			err = z[zb0004].Data.EncodeMsg(en)
			if err != nil {
				return
			}
		}
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z Transactions) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	o = msgp.AppendArrayHeader(o, uint32(len(z)))
	for zb0004 := range z {
		if z[zb0004] == nil {
			o = msgp.AppendNil(o)
		} else {
			// map header, size 1
			// string "Data"
			o = append(o, 0x81, 0xa4, 0x44, 0x61, 0x74, 0x61)
			o, err = z[zb0004].Data.MarshalMsg(o)
			if err != nil {
				return
			}
		}
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *Transactions) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var zb0002 uint32
	zb0002, bts, err = msgp.ReadArrayHeaderBytes(bts)
	if err != nil {
		return
	}
	if cap((*z)) >= int(zb0002) {
		(*z) = (*z)[:zb0002]
	} else {
		(*z) = make(Transactions, zb0002)
	}
	for zb0001 := range *z {
		if msgp.IsNil(bts) {
			bts, err = msgp.ReadNilBytes(bts)
			if err != nil {
				return
			}
			(*z)[zb0001] = nil
		} else {
			if (*z)[zb0001] == nil {
				(*z)[zb0001] = new(Transaction)
			}
			var field []byte
			_ = field
			var zb0003 uint32
			zb0003, bts, err = msgp.ReadMapHeaderBytes(bts)
			if err != nil {
				return
			}
			for zb0003 > 0 {
				zb0003--
				field, bts, err = msgp.ReadMapKeyZC(bts)
				if err != nil {
					return
				}
				switch msgp.UnsafeString(field) {
				case "Data":
					bts, err = (*z)[zb0001].Data.UnmarshalMsg(bts)
					if err != nil {
						return
					}
				default:
					bts, err = msgp.Skip(bts)
					if err != nil {
						return
					}
				}
			}
		}
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z Transactions) Msgsize() (s int) {
	s = msgp.ArrayHeaderSize
	for zb0004 := range z {
		if z[zb0004] == nil {
			s += msgp.NilSize
		} else {
			s += 1 + 5 + z[zb0004].Data.Msgsize()
		}
	}
	return
}

// DecodeMsg implements msgp.Decodable
func (z *TransactionsByPriorityAndNonce) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		default:
			err = dc.Skip()
			if err != nil {
				return
			}
		}
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z TransactionsByPriorityAndNonce) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 0
	err = en.Append(0x80)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z TransactionsByPriorityAndNonce) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 0
	o = append(o, 0x80)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *TransactionsByPriorityAndNonce) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		default:
			bts, err = msgp.Skip(bts)
			if err != nil {
				return
			}
		}
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z TransactionsByPriorityAndNonce) Msgsize() (s int) {
	s = 1
	return
}

// DecodeMsg implements msgp.Decodable
func (z *TxByNonce) DecodeMsg(dc *msgp.Reader) (err error) {
	var zb0002 uint32
	zb0002, err = dc.ReadArrayHeader()
	if err != nil {
		return
	}
	if cap((*z)) >= int(zb0002) {
		(*z) = (*z)[:zb0002]
	} else {
		(*z) = make(TxByNonce, zb0002)
	}
	for zb0001 := range *z {
		if dc.IsNil() {
			err = dc.ReadNil()
			if err != nil {
				return
			}
			(*z)[zb0001] = nil
		} else {
			if (*z)[zb0001] == nil {
				(*z)[zb0001] = new(Transaction)
			}
			var field []byte
			_ = field
			var zb0003 uint32
			zb0003, err = dc.ReadMapHeader()
			if err != nil {
				return
			}
			for zb0003 > 0 {
				zb0003--
				field, err = dc.ReadMapKeyPtr()
				if err != nil {
					return
				}
				switch msgp.UnsafeString(field) {
				case "Data":
					err = (*z)[zb0001].Data.DecodeMsg(dc)
					if err != nil {
						return
					}
				default:
					err = dc.Skip()
					if err != nil {
						return
					}
				}
			}
		}
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z TxByNonce) EncodeMsg(en *msgp.Writer) (err error) {
	err = en.WriteArrayHeader(uint32(len(z)))
	if err != nil {
		return
	}
	for zb0004 := range z {
		if z[zb0004] == nil {
			err = en.WriteNil()
			if err != nil {
				return
			}
		} else {
			// map header, size 1
			// write "Data"
			err = en.Append(0x81, 0xa4, 0x44, 0x61, 0x74, 0x61)
			if err != nil {
				return
			}
			err = z[zb0004].Data.EncodeMsg(en)
			if err != nil {
				return
			}
		}
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z TxByNonce) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	o = msgp.AppendArrayHeader(o, uint32(len(z)))
	for zb0004 := range z {
		if z[zb0004] == nil {
			o = msgp.AppendNil(o)
		} else {
			// map header, size 1
			// string "Data"
			o = append(o, 0x81, 0xa4, 0x44, 0x61, 0x74, 0x61)
			o, err = z[zb0004].Data.MarshalMsg(o)
			if err != nil {
				return
			}
		}
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *TxByNonce) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var zb0002 uint32
	zb0002, bts, err = msgp.ReadArrayHeaderBytes(bts)
	if err != nil {
		return
	}
	if cap((*z)) >= int(zb0002) {
		(*z) = (*z)[:zb0002]
	} else {
		(*z) = make(TxByNonce, zb0002)
	}
	for zb0001 := range *z {
		if msgp.IsNil(bts) {
			bts, err = msgp.ReadNilBytes(bts)
			if err != nil {
				return
			}
			(*z)[zb0001] = nil
		} else {
			if (*z)[zb0001] == nil {
				(*z)[zb0001] = new(Transaction)
			}
			var field []byte
			_ = field
			var zb0003 uint32
			zb0003, bts, err = msgp.ReadMapHeaderBytes(bts)
			if err != nil {
				return
			}
			for zb0003 > 0 {
				zb0003--
				field, bts, err = msgp.ReadMapKeyZC(bts)
				if err != nil {
					return
				}
				switch msgp.UnsafeString(field) {
				case "Data":
					bts, err = (*z)[zb0001].Data.UnmarshalMsg(bts)
					if err != nil {
						return
					}
				default:
					bts, err = msgp.Skip(bts)
					if err != nil {
						return
					}
				}
			}
		}
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z TxByNonce) Msgsize() (s int) {
	s = msgp.ArrayHeaderSize
	for zb0004 := range z {
		if z[zb0004] == nil {
			s += msgp.NilSize
		} else {
			s += 1 + 5 + z[zb0004].Data.Msgsize()
		}
	}
	return
}

// DecodeMsg implements msgp.Decodable
func (z *TxByPriority) DecodeMsg(dc *msgp.Reader) (err error) {
	var zb0002 uint32
	zb0002, err = dc.ReadArrayHeader()
	if err != nil {
		return
	}
	if cap((*z)) >= int(zb0002) {
		(*z) = (*z)[:zb0002]
	} else {
		(*z) = make(TxByPriority, zb0002)
	}
	for zb0001 := range *z {
		if dc.IsNil() {
			err = dc.ReadNil()
			if err != nil {
				return
			}
			(*z)[zb0001] = nil
		} else {
			if (*z)[zb0001] == nil {
				(*z)[zb0001] = new(Transaction)
			}
			var field []byte
			_ = field
			var zb0003 uint32
			zb0003, err = dc.ReadMapHeader()
			if err != nil {
				return
			}
			for zb0003 > 0 {
				zb0003--
				field, err = dc.ReadMapKeyPtr()
				if err != nil {
					return
				}
				switch msgp.UnsafeString(field) {
				case "Data":
					err = (*z)[zb0001].Data.DecodeMsg(dc)
					if err != nil {
						return
					}
				default:
					err = dc.Skip()
					if err != nil {
						return
					}
				}
			}
		}
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z TxByPriority) EncodeMsg(en *msgp.Writer) (err error) {
	err = en.WriteArrayHeader(uint32(len(z)))
	if err != nil {
		return
	}
	for zb0004 := range z {
		if z[zb0004] == nil {
			err = en.WriteNil()
			if err != nil {
				return
			}
		} else {
			// map header, size 1
			// write "Data"
			err = en.Append(0x81, 0xa4, 0x44, 0x61, 0x74, 0x61)
			if err != nil {
				return
			}
			err = z[zb0004].Data.EncodeMsg(en)
			if err != nil {
				return
			}
		}
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z TxByPriority) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	o = msgp.AppendArrayHeader(o, uint32(len(z)))
	for zb0004 := range z {
		if z[zb0004] == nil {
			o = msgp.AppendNil(o)
		} else {
			// map header, size 1
			// string "Data"
			o = append(o, 0x81, 0xa4, 0x44, 0x61, 0x74, 0x61)
			o, err = z[zb0004].Data.MarshalMsg(o)
			if err != nil {
				return
			}
		}
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *TxByPriority) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var zb0002 uint32
	zb0002, bts, err = msgp.ReadArrayHeaderBytes(bts)
	if err != nil {
		return
	}
	if cap((*z)) >= int(zb0002) {
		(*z) = (*z)[:zb0002]
	} else {
		(*z) = make(TxByPriority, zb0002)
	}
	for zb0001 := range *z {
		if msgp.IsNil(bts) {
			bts, err = msgp.ReadNilBytes(bts)
			if err != nil {
				return
			}
			(*z)[zb0001] = nil
		} else {
			if (*z)[zb0001] == nil {
				(*z)[zb0001] = new(Transaction)
			}
			var field []byte
			_ = field
			var zb0003 uint32
			zb0003, bts, err = msgp.ReadMapHeaderBytes(bts)
			if err != nil {
				return
			}
			for zb0003 > 0 {
				zb0003--
				field, bts, err = msgp.ReadMapKeyZC(bts)
				if err != nil {
					return
				}
				switch msgp.UnsafeString(field) {
				case "Data":
					bts, err = (*z)[zb0001].Data.UnmarshalMsg(bts)
					if err != nil {
						return
					}
				default:
					bts, err = msgp.Skip(bts)
					if err != nil {
						return
					}
				}
			}
		}
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z TxByPriority) Msgsize() (s int) {
	s = msgp.ArrayHeaderSize
	for zb0004 := range z {
		if z[zb0004] == nil {
			s += msgp.NilSize
		} else {
			s += 1 + 5 + z[zb0004].Data.Msgsize()
		}
	}
	return
}

// DecodeMsg implements msgp.Decodable
func (z *Txdata) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "AccountNonce":
			z.AccountNonce, err = dc.ReadUint64()
			if err != nil {
				return
			}
		case "Actions":
			var zb0002 uint32
			zb0002, err = dc.ReadArrayHeader()
			if err != nil {
				return
			}
			if cap(z.Actions) >= int(zb0002) {
				z.Actions = (z.Actions)[:zb0002]
			} else {
				z.Actions = make(ActionSlice, zb0002)
			}
			for za0001 := range z.Actions {
				var zb0003 uint32
				zb0003, err = dc.ReadMapHeader()
				if err != nil {
					return
				}
				for zb0003 > 0 {
					zb0003--
					field, err = dc.ReadMapKeyPtr()
					if err != nil {
						return
					}
					switch msgp.UnsafeString(field) {
					case "Address":
						if dc.IsNil() {
							err = dc.ReadNil()
							if err != nil {
								return
							}
							z.Actions[za0001].Address = nil
						} else {
							if z.Actions[za0001].Address == nil {
								z.Actions[za0001].Address = new(types.Address)
							}
							err = z.Actions[za0001].Address.DecodeMsg(dc)
							if err != nil {
								return
							}
						}
					case "Params":
						z.Actions[za0001].Params, err = dc.ReadBytes(z.Actions[za0001].Params)
						if err != nil {
							return
						}
					default:
						err = dc.Skip()
						if err != nil {
							return
						}
					}
				}
			}
		case "V":
			if dc.IsNil() {
				err = dc.ReadNil()
				if err != nil {
					return
				}
				z.V = nil
			} else {
				if z.V == nil {
					z.V = new(types.BigInt)
				}
				err = z.V.DecodeMsg(dc)
				if err != nil {
					return
				}
			}
		case "R":
			if dc.IsNil() {
				err = dc.ReadNil()
				if err != nil {
					return
				}
				z.R = nil
			} else {
				if z.R == nil {
					z.R = new(types.BigInt)
				}
				err = z.R.DecodeMsg(dc)
				if err != nil {
					return
				}
			}
		case "S":
			if dc.IsNil() {
				err = dc.ReadNil()
				if err != nil {
					return
				}
				z.S = nil
			} else {
				if z.S == nil {
					z.S = new(types.BigInt)
				}
				err = z.S.DecodeMsg(dc)
				if err != nil {
					return
				}
			}
		default:
			err = dc.Skip()
			if err != nil {
				return
			}
		}
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z *Txdata) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 5
	// write "AccountNonce"
	err = en.Append(0x85, 0xac, 0x41, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x4e, 0x6f, 0x6e, 0x63, 0x65)
	if err != nil {
		return
	}
	err = en.WriteUint64(z.AccountNonce)
	if err != nil {
		return
	}
	// write "Actions"
	err = en.Append(0xa7, 0x41, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73)
	if err != nil {
		return
	}
	err = en.WriteArrayHeader(uint32(len(z.Actions)))
	if err != nil {
		return
	}
	for za0001 := range z.Actions {
		// map header, size 2
		// write "Address"
		err = en.Append(0x82, 0xa7, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73)
		if err != nil {
			return
		}
		if z.Actions[za0001].Address == nil {
			err = en.WriteNil()
			if err != nil {
				return
			}
		} else {
			err = z.Actions[za0001].Address.EncodeMsg(en)
			if err != nil {
				return
			}
		}
		// write "Params"
		err = en.Append(0xa6, 0x50, 0x61, 0x72, 0x61, 0x6d, 0x73)
		if err != nil {
			return
		}
		err = en.WriteBytes(z.Actions[za0001].Params)
		if err != nil {
			return
		}
	}
	// write "V"
	err = en.Append(0xa1, 0x56)
	if err != nil {
		return
	}
	if z.V == nil {
		err = en.WriteNil()
		if err != nil {
			return
		}
	} else {
		err = z.V.EncodeMsg(en)
		if err != nil {
			return
		}
	}
	// write "R"
	err = en.Append(0xa1, 0x52)
	if err != nil {
		return
	}
	if z.R == nil {
		err = en.WriteNil()
		if err != nil {
			return
		}
	} else {
		err = z.R.EncodeMsg(en)
		if err != nil {
			return
		}
	}
	// write "S"
	err = en.Append(0xa1, 0x53)
	if err != nil {
		return
	}
	if z.S == nil {
		err = en.WriteNil()
		if err != nil {
			return
		}
	} else {
		err = z.S.EncodeMsg(en)
		if err != nil {
			return
		}
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *Txdata) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 5
	// string "AccountNonce"
	o = append(o, 0x85, 0xac, 0x41, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x4e, 0x6f, 0x6e, 0x63, 0x65)
	o = msgp.AppendUint64(o, z.AccountNonce)
	// string "Actions"
	o = append(o, 0xa7, 0x41, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73)
	o = msgp.AppendArrayHeader(o, uint32(len(z.Actions)))
	for za0001 := range z.Actions {
		// map header, size 2
		// string "Address"
		o = append(o, 0x82, 0xa7, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73)
		if z.Actions[za0001].Address == nil {
			o = msgp.AppendNil(o)
		} else {
			o, err = z.Actions[za0001].Address.MarshalMsg(o)
			if err != nil {
				return
			}
		}
		// string "Params"
		o = append(o, 0xa6, 0x50, 0x61, 0x72, 0x61, 0x6d, 0x73)
		o = msgp.AppendBytes(o, z.Actions[za0001].Params)
	}
	// string "V"
	o = append(o, 0xa1, 0x56)
	if z.V == nil {
		o = msgp.AppendNil(o)
	} else {
		o, err = z.V.MarshalMsg(o)
		if err != nil {
			return
		}
	}
	// string "R"
	o = append(o, 0xa1, 0x52)
	if z.R == nil {
		o = msgp.AppendNil(o)
	} else {
		o, err = z.R.MarshalMsg(o)
		if err != nil {
			return
		}
	}
	// string "S"
	o = append(o, 0xa1, 0x53)
	if z.S == nil {
		o = msgp.AppendNil(o)
	} else {
		o, err = z.S.MarshalMsg(o)
		if err != nil {
			return
		}
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *Txdata) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "AccountNonce":
			z.AccountNonce, bts, err = msgp.ReadUint64Bytes(bts)
			if err != nil {
				return
			}
		case "Actions":
			var zb0002 uint32
			zb0002, bts, err = msgp.ReadArrayHeaderBytes(bts)
			if err != nil {
				return
			}
			if cap(z.Actions) >= int(zb0002) {
				z.Actions = (z.Actions)[:zb0002]
			} else {
				z.Actions = make(ActionSlice, zb0002)
			}
			for za0001 := range z.Actions {
				var zb0003 uint32
				zb0003, bts, err = msgp.ReadMapHeaderBytes(bts)
				if err != nil {
					return
				}
				for zb0003 > 0 {
					zb0003--
					field, bts, err = msgp.ReadMapKeyZC(bts)
					if err != nil {
						return
					}
					switch msgp.UnsafeString(field) {
					case "Address":
						if msgp.IsNil(bts) {
							bts, err = msgp.ReadNilBytes(bts)
							if err != nil {
								return
							}
							z.Actions[za0001].Address = nil
						} else {
							if z.Actions[za0001].Address == nil {
								z.Actions[za0001].Address = new(types.Address)
							}
							bts, err = z.Actions[za0001].Address.UnmarshalMsg(bts)
							if err != nil {
								return
							}
						}
					case "Params":
						z.Actions[za0001].Params, bts, err = msgp.ReadBytesBytes(bts, z.Actions[za0001].Params)
						if err != nil {
							return
						}
					default:
						bts, err = msgp.Skip(bts)
						if err != nil {
							return
						}
					}
				}
			}
		case "V":
			if msgp.IsNil(bts) {
				bts, err = msgp.ReadNilBytes(bts)
				if err != nil {
					return
				}
				z.V = nil
			} else {
				if z.V == nil {
					z.V = new(types.BigInt)
				}
				bts, err = z.V.UnmarshalMsg(bts)
				if err != nil {
					return
				}
			}
		case "R":
			if msgp.IsNil(bts) {
				bts, err = msgp.ReadNilBytes(bts)
				if err != nil {
					return
				}
				z.R = nil
			} else {
				if z.R == nil {
					z.R = new(types.BigInt)
				}
				bts, err = z.R.UnmarshalMsg(bts)
				if err != nil {
					return
				}
			}
		case "S":
			if msgp.IsNil(bts) {
				bts, err = msgp.ReadNilBytes(bts)
				if err != nil {
					return
				}
				z.S = nil
			} else {
				if z.S == nil {
					z.S = new(types.BigInt)
				}
				bts, err = z.S.UnmarshalMsg(bts)
				if err != nil {
					return
				}
			}
		default:
			bts, err = msgp.Skip(bts)
			if err != nil {
				return
			}
		}
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z *Txdata) Msgsize() (s int) {
	s = 1 + 13 + msgp.Uint64Size + 8 + msgp.ArrayHeaderSize
	for za0001 := range z.Actions {
		s += 1 + 8
		if z.Actions[za0001].Address == nil {
			s += msgp.NilSize
		} else {
			s += z.Actions[za0001].Address.Msgsize()
		}
		s += 7 + msgp.BytesPrefixSize + len(z.Actions[za0001].Params)
	}
	s += 2
	if z.V == nil {
		s += msgp.NilSize
	} else {
		s += z.V.Msgsize()
	}
	s += 2
	if z.R == nil {
		s += msgp.NilSize
	} else {
		s += z.R.Msgsize()
	}
	s += 2
	if z.S == nil {
		s += msgp.NilSize
	} else {
		s += z.S.Msgsize()
	}
	return
}
