# CryptShell: Multi-Client C2 Simulation Framework

CryptShell, Go dili kullanılarak siber güvenlik eğitim ve simülasyon süreçleri için geliştirilmiş, hafif ve modüler bir **Komuta ve Kontrol (C2 - Command & Control)** prototipidir. Bu proje, bir saldırganın ağ üzerinden birden fazla kurban makineyi nasıl yönettiğini ve siber güvenlik uzmanlarının bu trafiği nasıl analiz ettiğini anlamak amacıyla tasarlanmış bir **MVP (Minimum Viable Product)** çalışmasıdır.

## 🚀 Öne Çıkan Özellikler

* **Çoklu İstemci Yönetimi (Multi-Client Concurrency):** Go dilinin eşzamanlılık (Goroutines ve Channels) yetenekleri sayesinde sunucu, çökmeden aynı anda birden fazla ajanı hafızasında tutabilir ve bağımsız olarak yönetebilir.
* **Trafik Şifreleme (Crypto Layer):** Ajan ile sunucu arasındaki tüm ağ iletişimi dinamik bir XOR algoritması ve Base64 katmanıyla şifrelenir. Bu sayede basit ağ izleme (Sniffing) araçları atlatılır.
* **Gelişmiş Dosya Operasyonları (Upload / Download):** Güvenli soket hatları üzerinden hedef sisteme uzaktan dosya enjekte etme (`upload`) ve hedef sistemden veri sızdırma (`download`) yetenekleri sıfırdan protokolize edilmiştir.
* **Gizlilik ve AV Evasion (Window Hiding):** İstemci tarafı, Windows GUI bayrakları (`-ldflags="-H windowsgui"`) kullanılarak derlendiğinde, kurban ekranında hiçbir konsol penceresi açılmadan tamamen arka plan süreci (Background Process) olarak sessizce çalışır.

## 🛠️ Mimari ve Komut Yapısı

Ana menü üzerinden ajanlar listelenebilir ve hedef seans seçilerek kurban makinenin işletim sisteminde (CMD) etkileşimli komutlar koşturulabilir.

* `sessions` -> Aktif bağlı olan tüm görünmez ajanları listeler.
* `interact <id>` -> Seçilen ajanın terminaline tünel kurar.
* `download <dosya_adı>` -> Kurban makineden sunucuya dosya sızdırır.
* `upload <dosya_adı>` -> Sunucudan kurban makineye dosya yükler.
* `back` -> Seansı kapatmadan arka plana alır ve ana menüye döner.
* `exit` -> Aktif ajanı hedef bellekten tamamen temizler (Kill process).

## 💻 Kurulum ve Çalıştırma

```markdown

1. Sunucuyu Başlatma
```bash
cd server
go run server.go
```

2. Ajanı Görünmez Modda Derleme ve Çalıştırma  📌 (Başlık ve Kod Bloğu Düzenlendi)
```bash
cd agent
go build -ldflags="-H windowsgui" agent.go
./agent.exe
```

⚠️ Yasal Uyarı / Disclaimer
Bu proje tamamen eğitim, savunma odaklı siber güvenlik simülasyonları ve araştırma amacıyla geliştirilmiştir. Yetkisiz sistemlerde kullanılması yasal sorumluluk doğurabilir. Tüm sorumluluk kullanıcıya aittir.
```
