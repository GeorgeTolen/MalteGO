# MalteGO

Go-реализация [greynoise-maltego](https://github.com/GreyNoise-Intelligence/greynoise-maltego) — сервер трансформов для Maltego, который подтягивает данные из GreyNoise API.

Если коротко: берёшь IP в Maltego, запускаешь трансформ — и граф разворачивается в акторов, CVE, похожие адреса, организации и всё что с этим IP связано. Мы переписали оригинальный Python/Flask сервис на Go с Gin.

---

## Что внутри

**13 трансформов в 4 группах:**

**Community** (бесплатно):
- `GreyNoiseCommunityIPLookup` — базовая проверка IP: шумит ли, известный ли сервис, malicious/benign

**Контекст по IP** (Enterprise):
- `GreyNoiseNoiseIPLookupAllDetails` — полный портрет: гео, ASN, теги, CVE, порты, актор
- `GreyNoiseNoiseIPLookupGetActor` — кто стоит за этим IP
- `GreyNoiseNoiseIPLookupGetCVEs` — какие уязвимости эксплуатирует
- `GreyNoiseNoiseIPLookupGetOrg` — провайдер и организация
- `GreyNoiseNoiseIPLookupGetPorts` — что сканирует
- `GreyNoiseNoiseIPLookupGetTags` — теги активности
- `GreyNoiseRIOTIPLookup` — это легитимный сервис? (убирает ложные алармы)

**Пивотинг** (Enterprise):
- `GreyNoiseNoiseIPSims` — похожие IP по поведению, строит всю инфраструктуру атакующего

**GNQL-запросы** (Enterprise):
- `GreyNoiseQueryByASN` — все вредоносные IP в сети
- `GreyNoiseQueryByActor` — все IP группировки
- `GreyNoiseQueryByCVE` — кто прямо сейчас эксплуатирует уязвимость
- `GreyNoiseQueryByTag` — IP по типу активности

---

## Стек

- **Go 1.26** + **Gin**
- Протокол **Maltego TRX** (XML request/response)
- **Docker** + docker-compose
- Конфигурация через `.env`

---

## Быстрый старт

```bash
git clone https://github.com/GeorgeTolen/MalteGO.git
cd MalteGO

cp .env.example .env
# Вставь GREYNOISE_API_KEY в .env (бесплатный ключ: https://viz.greynoise.io/signup)

make docker-run   # сборка + запуск в Docker
# или
make run          # локально без Docker
```

Сервис поднимается на `http://localhost:8080`.

---

## Проверить что работает

```bash
# Список трансформов
curl http://localhost:8080/

# Community lookup (работает без Enterprise-ключа)
curl -X POST http://localhost:8080/run/GreyNoiseCommunityIPLookup \
  -H "Content-Type: text/xml" \
  -d '<MaltegoMessage><MaltegoTransformRequestMessage>
    <Entities><Entity Type="maltego.IPv4Address">
      <Value>8.8.8.8</Value><Weight>100</Weight>
    </Entity></Entities>
    <TransformFields>
      <Field Name="greynoise.api.key">YOUR_KEY</Field>
    </TransformFields>
    <Limits SoftLimit="12" HardLimit="12"/>
  </MaltegoTransformRequestMessage></MaltegoMessage>'
```

---

## Make-команды

```
make run            # запустить локально
make docker-run     # собрать образ + запустить в Docker
make test           # тесты
make test-cover     # тесты с отчётом покрытия
make build          # собрать бинарник
make help           # все команды
```

---

## API-ключ GreyNoise

| Тир | Цена | Что доступно |
|-----|------|-------------|
| Community | Бесплатно | 1 трансформ, 50 запросов/неделю |
| Enterprise Trial | Бесплатно 14 дней | Все 13 трансформов |
| Enterprise | Платно | Все 13 трансформов, без лимита |

Получить ключ: [viz.greynoise.io/signup](https://viz.greynoise.io/signup)

Ключ можно передавать двумя способами — через `.env` или прямо в теле каждого запроса в `TransformFields`. Второй способ удобен при подключении из Maltego Desktop.

Подробная документация по всем эндпоинтам и примеры для Postman — в [API.md](API.md).
