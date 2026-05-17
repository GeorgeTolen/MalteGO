# MalteGO — API Reference & Testing Guide

## Запуск сервиса

```bash
make run          # локально, порт 8080
make docker-run   # Docker, порт 8080
```

Сервис всегда отвечает HTTP 200. Ошибки передаются внутри XML как `FatalError` —
так требует протокол Maltego TRX.

---

## Что работает прямо сейчас (без API-ключа / с бесплатным ключом)

### GET /
**Health-check.** Возвращает список зарегистрированных трансформов.

**Postman:**
```
GET http://localhost:8080/
```

**Ответ:**
```json
{
  "count": 13,
  "service": "MalteGO — GreyNoise Maltego Transform Server",
  "transforms": [
    "GreyNoiseCommunityIPLookup",
    "GreyNoiseNoiseIPLookupAllDetails",
    ...
  ]
}
```

---

### POST /run/GreyNoiseCommunityIPLookup ✅ РАБОТАЕТ БЕСПЛАТНО

Запрашивает Community API GreyNoise. Возвращает базовую информацию по IP:
является ли он сканером (`noise`), известным сервисом (`riot`), классификацию и имя актора.

**Лимит бесплатного ключа:** 50 запросов/неделю.  
**Без ключа:** 10 запросов/день (ключ всё равно нужно передать — можно любую строку).

**Postman:**
```
POST http://localhost:8080/run/GreyNoiseCommunityIPLookup
Content-Type: text/xml
```

**Body → raw → XML:**
```xml
<?xml version="1.0"?>
<MaltegoMessage>
  <MaltegoTransformRequestMessage>
    <Entities>
      <Entity Type="maltego.IPv4Address">
        <Value>8.8.8.8</Value>
        <Weight>100</Weight>
      </Entity>
    </Entities>
    <TransformFields>
      <Field Name="greynoise.api.key">ВАШ_КЛЮЧ_ИЛИ_ЛЮБАЯ_СТРОКА</Field>
    </TransformFields>
    <Limits SoftLimit="12" HardLimit="12"/>
  </MaltegoTransformRequestMessage>
</MaltegoMessage>
```

**Ответ (8.8.8.8 → Google Public DNS):**
```xml
<?xml version="1.0" encoding="UTF-8"?>
<MaltegoMessage>
  <MaltegoTransformResponseMessage>
    <Entities>
      <Entity Type="maltego.IPv4Address">
        <Value>8.8.8.8</Value>
        <Weight>100</Weight>
        <DisplayInformation>
          <Label Name="GreyNoise Community" Type="text/html">
            <![CDATA[<b>IP:</b> 8.8.8.8<br/>
            <b>Noise:</b> false<br/>
            <b>RIOT:</b> true<br/>
            <b>Classification:</b> benign<br/>
            <b>Actor:</b> Google Public DNS<br/>
            <b>Last Seen:</b> 2026-05-16]]>
          </Label>
        </DisplayInformation>
        <AdditionalFields>
          <Field Name="noise"           DisplayName="Noise"           MatchingRule="loose">false</Field>
          <Field Name="riot"            DisplayName="RIOT"            MatchingRule="loose">true</Field>
          <Field Name="classification"  DisplayName="Classification"  MatchingRule="loose">benign</Field>
          <Field Name="actor"           DisplayName="Actor"           MatchingRule="loose">Google Public DNS</Field>
          <Field Name="last_seen"       DisplayName="Last Seen"       MatchingRule="loose">2026-05-16</Field>
          <Field Name="link"            DisplayName="GreyNoise Link"  MatchingRule="loose">https://viz.greynoise.io/ip/8.8.8.8</Field>
        </AdditionalFields>
        <IconURL>https://www.greynoise.io/favicon.ico</IconURL>
      </Entity>
    </Entities>
    <UIMessages>
      <UIMessage MessageType="Inform">Community lookup for 8.8.8.8 complete</UIMessage>
    </UIMessages>
  </MaltegoTransformResponseMessage>
</MaltegoMessage>
```

**Интересные IP для тестов:**

| IP | Что вернёт |
|----|-----------|
| `8.8.8.8` | Google Public DNS, riot=true, benign |
| `1.1.1.1` | Cloudflare DNS, riot=true, benign |
| `185.220.101.1` | Tor exit node, noise=true, malicious |
| `45.83.64.1` | Известный сканер, noise=true |
| `192.168.1.1` | Приватный IP → "not observed" |

---

## Что будет работать с платным / 14-дневным trial ключом

Получить trial: [viz.greynoise.io/signup](https://viz.greynoise.io/signup) → запросить Enterprise Trial.

### POST /run/GreyNoiseNoiseIPLookupAllDetails 🔑 ENTERPRISE

Полный контекст по IP: geo, ASN, организация, теги, CVE, порты, актор, первое/последнее появление.

**Body:**
```xml
<?xml version="1.0"?>
<MaltegoMessage>
  <MaltegoTransformRequestMessage>
    <Entities>
      <Entity Type="maltego.IPv4Address">
        <Value>185.220.101.1</Value>
        <Weight>100</Weight>
      </Entity>
    </Entities>
    <TransformFields>
      <Field Name="greynoise.api.key">ВАШ_ENTERPRISE_КЛЮЧ</Field>
    </TransformFields>
    <Limits SoftLimit="12" HardLimit="12"/>
  </MaltegoTransformRequestMessage>
</MaltegoMessage>
```

**Что вернёт:** одна entity с полями classification, actor, first_seen, last_seen, asn, organization, country, tags, cves, spoofable, tor, os.

---

### POST /run/GreyNoiseNoiseIPLookupGetActor 🔑 ENTERPRISE

По IP → имя актора (threat actor). Возвращает entity типа `maltego.Person`.

**Body:** тот же XML, `<Value>IP_АДРЕС</Value>`.

**Что вернёт:**
```xml
<Entity Type="maltego.Person">
  <Value>Mirai</Value>
  ...
</Entity>
```

---

### POST /run/GreyNoiseNoiseIPLookupGetCVEs 🔑 ENTERPRISE

По IP → список CVE которые он эксплуатирует. Каждый CVE — отдельная entity типа `maltego.CVE`.

**Что вернёт:** N entities по числу CVE:
```xml
<Entity Type="maltego.CVE"><Value>CVE-2021-44228</Value></Entity>
<Entity Type="maltego.CVE"><Value>CVE-2022-0001</Value></Entity>
```

---

### POST /run/GreyNoiseNoiseIPLookupGetOrg 🔑 ENTERPRISE

По IP → организация/провайдер. Entity типа `maltego.Organization`.

**Что вернёт:**
```xml
<Entity Type="maltego.Organization">
  <Value>Acme Hosting Ltd</Value>
  <AdditionalFields>
    <Field Name="asn">AS12345</Field>
    <Field Name="country">RU</Field>
    <Field Name="category">hosting</Field>
  </AdditionalFields>
</Entity>
```

---

### POST /run/GreyNoiseNoiseIPLookupGetPorts 🔑 ENTERPRISE

По IP → открытые порты (raw scan data). Каждый порт — entity типа `maltego.Port`.

**Что вернёт:**
```xml
<Entity Type="maltego.Port"><Value>22</Value>...</Entity>
<Entity Type="maltego.Port"><Value>80</Value>...</Entity>
<Entity Type="maltego.Port"><Value>443</Value>...</Entity>
```

---

### POST /run/GreyNoiseNoiseIPLookupGetTags 🔑 ENTERPRISE

По IP → теги активности. Каждый тег — entity типа `maltego.Hashtag`.

**Что вернёт:**
```xml
<Entity Type="maltego.Hashtag"><Value>mirai</Value></Entity>
<Entity Type="maltego.Hashtag"><Value>scanner</Value></Entity>
```

---

### POST /run/GreyNoiseRIOTIPLookup 🔑 ENTERPRISE

RIOT = Rule It Out. Определяет, является ли IP известным легитимным сервисом
(CDN, DNS, облачный провайдер). Помогает убрать ложные алармы.

**Что вернёт (если riot=true):**
```xml
<Entity Type="maltego.IPv4Address">
  <Value>8.8.8.8</Value>
  <AdditionalFields>
    <Field Name="riot">true</Field>
    <Field Name="name">Google Public DNS</Field>
    <Field Name="category">public_dns</Field>
    <Field Name="trust_level">1</Field>
    <Field Name="description">Google DNS resolver</Field>
  </AdditionalFields>
</Entity>
```

---

### POST /run/GreyNoiseNoiseIPSims 🔑 ENTERPRISE

По IP → похожие IP (схожее поведение, теги, активность). Пивотинг от одного IP к целой сети.

**Параметр `HardLimit`** управляет количеством результатов:
```xml
<Limits SoftLimit="12" HardLimit="50"/>
```

**Что вернёт:** N entities типа `maltego.IPv4Address`:
```xml
<Entity Type="maltego.IPv4Address">
  <Value>5.5.5.5</Value>
  <AdditionalFields>
    <Field Name="similarity_score">0.9500</Field>
    <Field Name="actor">Mirai</Field>
    <Field Name="features">tags,ports,asn</Field>
  </AdditionalFields>
</Entity>
```

---

### POST /run/GreyNoiseQueryByASN 🔑 ENTERPRISE

По ASN-номеру → все IP в этой сети, замеченные GreyNoise.  
Принимает `AS15169` или просто `15169` — сервис добавит префикс `AS` автоматически.

**Body:**
```xml
<Entity Type="maltego.AS">
  <Value>AS15169</Value>
</Entity>
```

**Что вернёт:** N entities `maltego.IPv4Address` с classification, actor, last_seen.

---

### POST /run/GreyNoiseQueryByActor 🔑 ENTERPRISE

По имени актора → все IP, приписанные этому актору.

**Body:**
```xml
<Entity Type="maltego.Person">
  <Value>Mirai</Value>
</Entity>
```

---

### POST /run/GreyNoiseQueryByCVE 🔑 ENTERPRISE

По CVE-ID → IP-адреса, которые эксплуатируют эту уязвимость.  
Принимает `CVE-2021-44228` или `cve-2021-44228` — сервис нормализует к uppercase.

**Body:**
```xml
<Entity Type="maltego.CVE">
  <Value>CVE-2021-44228</Value>
</Entity>
```

---

### POST /run/GreyNoiseQueryByTag 🔑 ENTERPRISE

По тегу → все IP с этим тегом.

**Body:**
```xml
<Entity Type="maltego.Hashtag">
  <Value>mirai</Value>
</Entity>
```

**Популярные теги для тестов:** `mirai`, `scanner`, `rdp-scanner`, `smb-scanner`, `zmap`

---

## Протокол: обработка ошибок

Все ошибки возвращаются HTTP 200 (требование Maltego TRX) с FatalError внутри XML:

| Ситуация | UIMessage |
|----------|-----------|
| Нет API-ключа | `GreyNoise API key not configured...` |
| Неверный ключ (401) | `api error 401: {"message":"unauthorized"}` |
| Неизвестный трансформ | `Transform "X" not found` |
| Пустой IP | `Entity value is empty` |
| Битый XML | `Invalid Maltego XML: xml parse: ...` |
| IP не в базе GreyNoise | `Inform: X has not been seen by GreyNoise` |

---

## Быстрая настройка в Postman

1. `New Collection` → назвать **MalteGO**
2. Добавить переменную коллекции: `base_url = http://localhost:8080`, `api_key = ВАШ_КЛЮЧ`
3. Для каждого трансформа создать запрос:
   - Method: `POST`
   - URL: `{{base_url}}/run/GreyNoiseCommunityIPLookup`
   - Headers: `Content-Type: text/xml`
   - Body → raw → XML (шаблон выше, заменить `{{api_key}}`)
4. `GET {{base_url}}/` — проверить что сервис живой

---

## Итоговая таблица трансформов

| Трансформ | Вход | Выход | Ключ |
|-----------|------|-------|------|
| `GreyNoiseCommunityIPLookup` | IPv4 | noise, riot, classification, actor | Free |
| `GreyNoiseNoiseIPLookupAllDetails` | IPv4 | полный контекст | Enterprise |
| `GreyNoiseNoiseIPLookupGetActor` | IPv4 | `maltego.Person` | Enterprise |
| `GreyNoiseNoiseIPLookupGetCVEs` | IPv4 | `maltego.CVE` × N | Enterprise |
| `GreyNoiseNoiseIPLookupGetOrg` | IPv4 | `maltego.Organization` | Enterprise |
| `GreyNoiseNoiseIPLookupGetPorts` | IPv4 | `maltego.Port` × N | Enterprise |
| `GreyNoiseNoiseIPLookupGetTags` | IPv4 | `maltego.Hashtag` × N | Enterprise |
| `GreyNoiseRIOTIPLookup` | IPv4 | riot + сервис | Enterprise |
| `GreyNoiseNoiseIPSims` | IPv4 | похожие IPv4 × N | Enterprise |
| `GreyNoiseQueryByASN` | AS-номер | IPv4 × N | Enterprise |
| `GreyNoiseQueryByActor` | имя актора | IPv4 × N | Enterprise |
| `GreyNoiseQueryByCVE` | CVE-ID | IPv4 × N | Enterprise |
| `GreyNoiseQueryByTag` | тег | IPv4 × N | Enterprise |
