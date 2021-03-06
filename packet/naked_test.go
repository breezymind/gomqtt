package packet

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNakedPacketDecode(t *testing.T) {
	pktBytes := []byte{
		byte(DISCONNECT << 4),
		0,
	}

	n, err := nakedPacketDecode(pktBytes, DISCONNECT)

	assert.NoError(t, err)
	assert.Equal(t, 2, n)
}

func TestNakedPacketDecodeError1(t *testing.T) {
	pktBytes := []byte{
		byte(DISCONNECT << 4),
		1, // < wrong remaining length
		0,
	}

	n, err := nakedPacketDecode(pktBytes, DISCONNECT)

	assert.Error(t, err)
	assert.Equal(t, 2, n)
}

func TestNakedPacketEncode(t *testing.T) {
	pktBytes := []byte{
		byte(DISCONNECT << 4),
		0,
	}

	dst := make([]byte, nakedPacketLen())
	n, err := nakedPacketEncode(dst, DISCONNECT)

	assert.NoError(t, err)
	assert.Equal(t, 2, n)
	assert.Equal(t, pktBytes, dst[:n])
}

func TestNakedPacketEqualDecodeEncode(t *testing.T) {
	pktBytes := []byte{
		byte(DISCONNECT << 4),
		0,
	}

	n, err := nakedPacketDecode(pktBytes, DISCONNECT)

	assert.NoError(t, err)
	assert.Equal(t, 2, n)

	dst := make([]byte, nakedPacketLen())
	n2, err := nakedPacketEncode(dst, DISCONNECT)

	assert.NoError(t, err)
	assert.Equal(t, 2, n2)
	assert.Equal(t, pktBytes, dst[:n2])

	n3, err := nakedPacketDecode(dst, DISCONNECT)

	assert.NoError(t, err)
	assert.Equal(t, 2, n3)
}

func BenchmarkNakedPacketEncode(b *testing.B) {
	buf := make([]byte, nakedPacketLen())

	for i := 0; i < b.N; i++ {
		_, err := nakedPacketEncode(buf, DISCONNECT)
		if err != nil {
			panic(err)
		}
	}
}

func BenchmarkNakedPacketDecode(b *testing.B) {
	pktBytes := []byte{
		byte(DISCONNECT << 4),
		0,
	}

	for i := 0; i < b.N; i++ {
		_, err := nakedPacketDecode(pktBytes, DISCONNECT)
		if err != nil {
			panic(err)
		}
	}
}

func testNakedPacketImplementation(t *testing.T, _t Type) {
	pkt, err := _t.New()
	assert.NoError(t, err)
	assert.Equal(t, _t, pkt.Type())
	assert.Equal(t, fmt.Sprintf("<%sPacket>", pkt.Type().String()), pkt.String())

	buf := make([]byte, pkt.Len())
	n, err := pkt.Encode(buf)
	assert.NoError(t, err)
	assert.Equal(t, 2, n)

	n, err = pkt.Decode(buf)
	assert.NoError(t, err)
	assert.Equal(t, 2, n)
}

func TestDisconnectImplementation(t *testing.T) {
	testNakedPacketImplementation(t, DISCONNECT)
}

func TestPingreqImplementation(t *testing.T) {
	testNakedPacketImplementation(t, PINGREQ)
}

func TestPingrespImplementation(t *testing.T) {
	testNakedPacketImplementation(t, PINGRESP)
}
