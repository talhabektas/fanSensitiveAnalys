# Database Scripts

Bu dizinde veritabanı bakımı için script'ler bulunmaktadır.

## Scripts

### 1. Turkish Characters Fix (`fix-turkish-chars/`)
Bozuk Türkçe karakterleri düzeltir.

```bash
cd scripts/fix-turkish-chars
go run main.go
```

### 2. Team Names Fix (`fix-team-names/`)
Takım isimlerindeki bozuk karakterleri düzeltir.

```bash
cd scripts/fix-team-names  
go run main.go
```

## Gereksinimler

- Go 1.21+
- MongoDB bağlantısı (`.env` dosyasında yapılandırılmış)
- Backend `config` modülüne erişim

## Notlar

- Her script ayrı bir modül olarak yapılandırılmıştır
- Ana backend modülüne `replace` direktifi ile bağlanırlar
- Production veritabanında çalıştırmadan önce backup alın