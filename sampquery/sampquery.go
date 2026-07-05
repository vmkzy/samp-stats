// Package sampquery реализует SA-MP Query Protocol.
// Документация протокола: https://sampwiki.blast.hk/wiki/Query_Mechanism
package sampquery

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"time"
)

type ServerInfo struct {
	Password   bool
	Players    uint16
	MaxPlayers uint16
	Hostname   string
	Gamemode   string
	Language   string
}

func buildPacket(ip string, port uint16, opcode byte) ([]byte, error) {
	parts := strings.Split(ip, ".")
	if len(parts) != 4 {
		return nil, fmt.Errorf("некорректный IP: %s", ip)
	}

	buf := new(bytes.Buffer)
	buf.WriteString("SAMP")

	for _, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil || n < 0 || n > 255 {
			return nil, fmt.Errorf("некорректный октет IP: %s", p)
		}
		buf.WriteByte(byte(n))
	}

	if err := binary.Write(buf, binary.LittleEndian, port); err != nil {
		return nil, err
	}

	buf.WriteByte(opcode)
	return buf.Bytes(), nil
}

func sendAndReceive(ip string, port uint16, packet []byte, timeout time.Duration) ([]byte, error) {
	addr := fmt.Sprintf("%s:%d", ip, port)
	conn, err := net.Dial("udp", addr)
	if err != nil {
		return nil, fmt.Errorf("не удалось подключиться: %w", err)
	}
	defer conn.Close()

	conn.SetDeadline(time.Now().Add(timeout))

	if _, err := conn.Write(packet); err != nil {
		return nil, fmt.Errorf("ошибка отправки: %w", err)
	}

	buf := make([]byte, 4096)
	n, err := conn.Read(buf)
	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			return nil, fmt.Errorf("сервер %s не отвечает (timeout)", addr)
		}
		return nil, fmt.Errorf("ошибка получения: %w", err)
	}

	return buf[:n], nil
}

func readPascalString32(data []byte, offset int) (string, int, error) {
	if offset+4 > len(data) {
		return "", offset, fmt.Errorf("недостаточно данных для чтения длины строки")
	}
	length := binary.LittleEndian.Uint32(data[offset : offset+4])
	offset += 4

	if offset+int(length) > len(data) {
		return "", offset, fmt.Errorf("недостаточно данных для чтения строки")
	}
	value := string(data[offset : offset+int(length)])
	offset += int(length)

	return value, offset, nil
}

func GetInfo(ip string, port uint16) (*ServerInfo, error) {
	packet, err := buildPacket(ip, port, 'i')
	if err != nil {
		return nil, err
	}

	data, err := sendAndReceive(ip, port, packet, 2*time.Second)
	if err != nil {
		return nil, err
	}

	offset := 11

	if offset+5 > len(data) {
		return nil, fmt.Errorf("ответ сервера слишком короткий")
	}

	password := data[offset] != 0
	offset++

	players := binary.LittleEndian.Uint16(data[offset : offset+2])
	offset += 2

	maxPlayers := binary.LittleEndian.Uint16(data[offset : offset+2])
	offset += 2

	hostname, offset, err := readPascalString32(data, offset)
	if err != nil {
		return nil, err
	}

	gamemode, offset, err := readPascalString32(data, offset)
	if err != nil {
		return nil, err
	}

	language, _, err := readPascalString32(data, offset)
	if err != nil {
		return nil, err
	}

	return &ServerInfo{
		Password:   password,
		Players:    players,
		MaxPlayers: maxPlayers,
		Hostname:   hostname,
		Gamemode:   gamemode,
		Language:   language,
	}, nil
}

func GetPing(ip string, port uint16) (time.Duration, error) {
	packet, err := buildPacket(ip, port, 'p')
	if err != nil {
		return 0, err
	}

	token := make([]byte, 4)
	rand.Read(token)
	packet = append(packet, token...)

	start := time.Now()
	data, err := sendAndReceive(ip, port, packet, 3*time.Second)
	if err != nil {
		return 0, err
	}
	elapsed := time.Since(start)

	if len(data) < 4 || !bytes.Equal(data[len(data)-4:], token) {
		return 0, fmt.Errorf("сервер вернул некорректный ping-токен")
	}

	return elapsed, nil
}
