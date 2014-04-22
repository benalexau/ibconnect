-- +goose Up

-- Guidelines:
-- 1. All timestamp columns store UTC *only*.
-- 2. Append-only tables are the default unless otherwise specified.
-- 3. SQL scripts (not app tier) inserts all mandatory default reference data.

-- monetary stores an amount in a currency's minor unit (eg cents), with the
-- iso_4217.minor_unit column defining the number of minor units per major unit.
CREATE TYPE monetary AS (
    iso_4217_code SMALLINT,
    amount BIGINT
);

-- iso_4217 uses the ISO-assigned SMALLINT as a PK given iso_4217_codes are designed
-- for machine identification and it enables easier interpretation of monetary values
-- without needing to perform an installation-specific iso_4217 foreign key lookup.
CREATE TABLE iso_4217 (
    iso_4217_code SMALLINT PRIMARY KEY,
    minor_unit SMALLINT NOT NULL,
    alphabetic_code CHAR(3) NOT NULL UNIQUE,
    currency VARCHAR(100) NOT NULL
);

CREATE TABLE account_type (
    id BIGSERIAL PRIMARY KEY,
    type_desc VARCHAR(100) NOT NULL UNIQUE
);

CREATE TABLE account (
    id BIGSERIAL PRIMARY KEY,
    account_code VARCHAR(20) NOT NULL UNIQUE
);

CREATE TABLE account_snapshot (
    id BIGSERIAL PRIMARY KEY,
    account_id BIGSERIAL NOT NULL REFERENCES account(id) ON DELETE RESTRICT,
    created TIMESTAMP NOT NULL,
    UNIQUE(account_id, created)
);

CREATE VIEW v_account_snapshot_latest AS (
    SELECT
        account_code, max(created) AS latest
    FROM
        account_snapshot,
	account
    WHERE
        account.id = account_snapshot.account_id
    GROUP BY account_code
);

CREATE TABLE account_amount (
    id BIGSERIAL PRIMARY KEY,
    account_snapshot_id BIGSERIAL NOT NULL REFERENCES account_snapshot(id) ON DELETE RESTRICT,
    account_type_id BIGSERIAL NOT NULL REFERENCES account_type(id) ON DELETE RESTRICT,
    cushion NUMERIC NOT NULL,
    look_ahead_next_change SMALLINT NOT NULL,
    accrued_cash monetary NOT NULL,
    available_funds monetary NOT NULL,
    buying_power monetary NOT NULL,
    equity_with_loan_value monetary NOT NULL,
    excess_liquidity monetary NOT NULL,
    full_available_funds monetary NOT NULL,
    full_excess_liquidity monetary NOT NULL,
    full_init_margin_req monetary NOT NULL,
    full_maint_margin_req monetary NOT NULL,
    gross_position_value monetary NOT NULL,
    init_margin_req monetary NOT NULL,
    look_ahead_available_funds monetary NOT NULL,
    look_ahead_excess_liquidity monetary NOT NULL,
    look_ahead_init_margin_req monetary NOT NULL,
    look_ahead_maint_margin_req monetary NOT NULL,
    maint_margin_req monetary NOT NULL,
    net_liquidation monetary NOT NULL,
    total_cash_balance monetary NOT NULL,
    total_cash_value monetary NOT NULL,
    UNIQUE(account_snapshot_id)
);

CREATE TABLE security_type (
    id BIGSERIAL PRIMARY KEY,
    security_type VARCHAR(100) NOT NULL UNIQUE
);

CREATE TABLE symbol (
    id BIGSERIAL PRIMARY KEY,
    symbol VARCHAR(100) NOT NULL UNIQUE
);

CREATE TABLE exchange (
    id BIGSERIAL PRIMARY KEY,
    exchange VARCHAR(100) NOT NULL UNIQUE
);

CREATE TABLE contract (
    id BIGSERIAL PRIMARY KEY,
    created TIMESTAMP NOT NULL,
    ib_contract_id BIGINT NOT NULL,
    iso_4217_code SMALLINT NOT NULL,
    symbol_id BIGSERIAL NOT NULL REFERENCES symbol(id) ON DELETE RESTRICT,
    local_symbol_id BIGSERIAL NOT NULL REFERENCES symbol(id) ON DELETE RESTRICT,
    security_type_id BIGSERIAL NOT NULL REFERENCES security_type(id) ON DELETE RESTRICT,
    primary_exchange_id BIGSERIAL NOT NULL REFERENCES exchange(id) ON DELETE RESTRICT
);

CREATE VIEW v_contract AS (
    SELECT
        contract.id AS contract_id,
        ib_contract_id, contract.iso_4217_code, currency, security_type, exchange,
	s.symbol,
        ls.symbol AS local_symbol
    FROM
        contract,
        iso_4217,
        security_type,
        exchange,
        symbol AS s,
        symbol AS ls
    WHERE
        iso_4217.iso_4217_code = contract.iso_4217_code AND
        security_type.id = contract.security_type_id AND
        exchange.id = contract.primary_exchange_id AND
        s.id = contract.symbol_id AND
        ls.id = contract.local_symbol_id
);

CREATE TABLE account_position (
    id BIGSERIAL PRIMARY KEY,
    account_snapshot_id BIGSERIAL NOT NULL REFERENCES account_snapshot(id) ON DELETE RESTRICT,
    contract_id BIGSERIAL NOT NULL REFERENCES contract(id) ON DELETE RESTRICT,
    pos BIGINT NOT NULL,
    market_price NUMERIC NOT NULL,
    market_value NUMERIC NOT NULL,
    average_cost NUMERIC NOT NULL,
    unrealized_pnl NUMERIC NOT NULL,
    realized_pnl NUMERIC NOT NULL,
    UNIQUE(account_snapshot_id, contract_id)
);

CREATE VIEW v_account_position AS (
    SELECT
        account_snapshot.id AS account_snapshot_id, created, account_code, pos,
	market_price, market_value, average_cost, unrealized_pnl, realized_pnl,
	-- start of v_contract
	ib_contract_id, iso_4217_code, currency, security_type, exchange,
        symbol, local_symbol
	-- end of v_contract
    FROM
        account_position,
	account_snapshot,
	account,
        v_contract
    WHERE
        account_snapshot.id = account_position.account_snapshot_id AND
        account.id = account_snapshot.account_id AND
        v_contract.contract_id = account_position.contract_id
    ORDER BY market_value
);

INSERT INTO iso_4217 VALUES (008, 2, 'ALL', 'Lek');
INSERT INTO iso_4217 VALUES (012, 2, 'DZD', 'Algerian Dinar');
INSERT INTO iso_4217 VALUES (032, 2, 'ARS', 'Argentine Peso');
INSERT INTO iso_4217 VALUES (036, 2, 'AUD', 'Australian Dollar');
INSERT INTO iso_4217 VALUES (044, 2, 'BSD', 'Bahamian Dollar');
INSERT INTO iso_4217 VALUES (048, 3, 'BHD', 'Bahraini Dinar');
INSERT INTO iso_4217 VALUES (050, 2, 'BDT', 'Taka');
INSERT INTO iso_4217 VALUES (051, 2, 'AMD', 'Armenian Dram');
INSERT INTO iso_4217 VALUES (052, 2, 'BBD', 'Barbados Dollar');
INSERT INTO iso_4217 VALUES (060, 2, 'BMD', 'Bermudian Dollar');
INSERT INTO iso_4217 VALUES (064, 2, 'BTN', 'Ngultrum');
INSERT INTO iso_4217 VALUES (068, 2, 'BOB', 'Boliviano');
INSERT INTO iso_4217 VALUES (072, 2, 'BWP', 'Pula');
INSERT INTO iso_4217 VALUES (084, 2, 'BZD', 'Belize Dollar');
INSERT INTO iso_4217 VALUES (090, 2, 'SBD', 'Solomon Islands Dollar');
INSERT INTO iso_4217 VALUES (096, 2, 'BND', 'Brunei Dollar');
INSERT INTO iso_4217 VALUES (104, 2, 'MMK', 'Kyat');
INSERT INTO iso_4217 VALUES (108, 0, 'BIF', 'Burundi Franc');
INSERT INTO iso_4217 VALUES (116, 2, 'KHR', 'Riel');
INSERT INTO iso_4217 VALUES (124, 2, 'CAD', 'Canadian Dollar');
INSERT INTO iso_4217 VALUES (132, 2, 'CVE', 'Cape Verde Escudo');
INSERT INTO iso_4217 VALUES (136, 2, 'KYD', 'Cayman Islands Dollar');
INSERT INTO iso_4217 VALUES (144, 2, 'LKR', 'Sri Lanka Rupee');
INSERT INTO iso_4217 VALUES (152, 0, 'CLP', 'Chilean Peso');
INSERT INTO iso_4217 VALUES (156, 2, 'CNY', 'Yuan Renminbi');
INSERT INTO iso_4217 VALUES (170, 2, 'COP', 'Colombian Peso');
INSERT INTO iso_4217 VALUES (174, 0, 'KMF', 'Comoro Franc');
INSERT INTO iso_4217 VALUES (188, 2, 'CRC', 'Costa Rican Colon');
INSERT INTO iso_4217 VALUES (191, 2, 'HRK', 'Croatian Kuna');
INSERT INTO iso_4217 VALUES (192, 2, 'CUP', 'Cuban Peso');
INSERT INTO iso_4217 VALUES (203, 2, 'CZK', 'Czech Koruna');
INSERT INTO iso_4217 VALUES (208, 2, 'DKK', 'Danish Krone');
INSERT INTO iso_4217 VALUES (214, 2, 'DOP', 'Dominican Peso');
INSERT INTO iso_4217 VALUES (222, 2, 'SVC', 'El Salvador Colon');
INSERT INTO iso_4217 VALUES (230, 2, 'ETB', 'Ethiopian Birr');
INSERT INTO iso_4217 VALUES (232, 2, 'ERN', 'Nakfa');
INSERT INTO iso_4217 VALUES (238, 2, 'FKP', 'Falkland Islands Pound');
INSERT INTO iso_4217 VALUES (242, 2, 'FJD', 'Fiji Dollar');
INSERT INTO iso_4217 VALUES (262, 0, 'DJF', 'Djibouti Franc');
INSERT INTO iso_4217 VALUES (270, 2, 'GMD', 'Dalasi');
INSERT INTO iso_4217 VALUES (292, 2, 'GIP', 'Gibraltar Pound');
INSERT INTO iso_4217 VALUES (320, 2, 'GTQ', 'Quetzal');
INSERT INTO iso_4217 VALUES (324, 0, 'GNF', 'Guinea Franc');
INSERT INTO iso_4217 VALUES (328, 2, 'GYD', 'Guyana Dollar');
INSERT INTO iso_4217 VALUES (332, 2, 'HTG', 'Gourde');
INSERT INTO iso_4217 VALUES (340, 2, 'HNL', 'Lempira');
INSERT INTO iso_4217 VALUES (344, 2, 'HKD', 'Hong Kong Dollar');
INSERT INTO iso_4217 VALUES (348, 2, 'HUF', 'Forint');
INSERT INTO iso_4217 VALUES (352, 0, 'ISK', 'Iceland Krona');
INSERT INTO iso_4217 VALUES (356, 2, 'INR', 'Indian Rupee');
INSERT INTO iso_4217 VALUES (360, 2, 'IDR', 'Rupiah');
INSERT INTO iso_4217 VALUES (364, 2, 'IRR', 'Iranian Rial');
INSERT INTO iso_4217 VALUES (368, 3, 'IQD', 'Iraqi Dinar');
INSERT INTO iso_4217 VALUES (376, 2, 'ILS', 'New Israeli Sheqel');
INSERT INTO iso_4217 VALUES (388, 2, 'JMD', 'Jamaican Dollar');
INSERT INTO iso_4217 VALUES (392, 0, 'JPY', 'Yen');
INSERT INTO iso_4217 VALUES (398, 2, 'KZT', 'Tenge');
INSERT INTO iso_4217 VALUES (400, 3, 'JOD', 'Jordanian Dinar');
INSERT INTO iso_4217 VALUES (404, 2, 'KES', 'Kenyan Shilling');
INSERT INTO iso_4217 VALUES (408, 2, 'KPW', 'North Korean Won');
INSERT INTO iso_4217 VALUES (410, 0, 'KRW', 'Won');
INSERT INTO iso_4217 VALUES (414, 3, 'KWD', 'Kuwaiti Dinar');
INSERT INTO iso_4217 VALUES (417, 2, 'KGS', 'Som');
INSERT INTO iso_4217 VALUES (418, 2, 'LAK', 'Kip');
INSERT INTO iso_4217 VALUES (422, 2, 'LBP', 'Lebanese Pound');
INSERT INTO iso_4217 VALUES (426, 2, 'LSL', 'Loti');
INSERT INTO iso_4217 VALUES (430, 2, 'LRD', 'Liberian Dollar');
INSERT INTO iso_4217 VALUES (434, 3, 'LYD', 'Libyan Dinar');
INSERT INTO iso_4217 VALUES (440, 2, 'LTL', 'Lithuanian Litas');
INSERT INTO iso_4217 VALUES (446, 2, 'MOP', 'Pataca');
INSERT INTO iso_4217 VALUES (454, 2, 'MWK', 'Kwacha');
INSERT INTO iso_4217 VALUES (458, 2, 'MYR', 'Malaysian Ringgit');
INSERT INTO iso_4217 VALUES (462, 2, 'MVR', 'Rufiyaa');
INSERT INTO iso_4217 VALUES (478, 2, 'MRO', 'Ouguiya');
INSERT INTO iso_4217 VALUES (480, 2, 'MUR', 'Mauritius Rupee');
INSERT INTO iso_4217 VALUES (484, 2, 'MXN', 'Mexican Peso');
INSERT INTO iso_4217 VALUES (496, 2, 'MNT', 'Tugrik');
INSERT INTO iso_4217 VALUES (498, 2, 'MDL', 'Moldovan Leu');
INSERT INTO iso_4217 VALUES (504, 2, 'MAD', 'Moroccan Dirham');
INSERT INTO iso_4217 VALUES (512, 3, 'OMR', 'Rial Omani');
INSERT INTO iso_4217 VALUES (516, 2, 'NAD', 'Namibia Dollar');
INSERT INTO iso_4217 VALUES (524, 2, 'NPR', 'Nepalese Rupee');
INSERT INTO iso_4217 VALUES (532, 2, 'ANG', 'Netherlands Antillean Guilder');
INSERT INTO iso_4217 VALUES (533, 2, 'AWG', 'Aruban Florin');
INSERT INTO iso_4217 VALUES (548, 0, 'VUV', 'Vatu');
INSERT INTO iso_4217 VALUES (554, 2, 'NZD', 'New Zealand Dollar');
INSERT INTO iso_4217 VALUES (558, 2, 'NIO', 'Cordoba Oro');
INSERT INTO iso_4217 VALUES (566, 2, 'NGN', 'Naira');
INSERT INTO iso_4217 VALUES (578, 2, 'NOK', 'Norwegian Krone');
INSERT INTO iso_4217 VALUES (586, 2, 'PKR', 'Pakistan Rupee');
INSERT INTO iso_4217 VALUES (590, 2, 'PAB', 'Balboa');
INSERT INTO iso_4217 VALUES (598, 2, 'PGK', 'Kina');
INSERT INTO iso_4217 VALUES (600, 0, 'PYG', 'Guarani');
INSERT INTO iso_4217 VALUES (604, 2, 'PEN', 'Nuevo Sol');
INSERT INTO iso_4217 VALUES (608, 2, 'PHP', 'Philippine Peso');
INSERT INTO iso_4217 VALUES (634, 2, 'QAR', 'Qatari Rial');
INSERT INTO iso_4217 VALUES (643, 2, 'RUB', 'Russian Ruble');
INSERT INTO iso_4217 VALUES (646, 0, 'RWF', 'Rwanda Franc');
INSERT INTO iso_4217 VALUES (654, 2, 'SHP', 'Saint Helena Pound');
INSERT INTO iso_4217 VALUES (678, 2, 'STD', 'Dobra');
INSERT INTO iso_4217 VALUES (682, 2, 'SAR', 'Saudi Riyal');
INSERT INTO iso_4217 VALUES (690, 2, 'SCR', 'Seychelles Rupee');
INSERT INTO iso_4217 VALUES (694, 2, 'SLL', 'Leone');
INSERT INTO iso_4217 VALUES (702, 2, 'SGD', 'Singapore Dollar');
INSERT INTO iso_4217 VALUES (704, 0, 'VND', 'Dong');
INSERT INTO iso_4217 VALUES (706, 2, 'SOS', 'Somali Shilling');
INSERT INTO iso_4217 VALUES (710, 2, 'ZAR', 'Rand');
INSERT INTO iso_4217 VALUES (728, 2, 'SSP', 'South Sudanese Pound');
INSERT INTO iso_4217 VALUES (748, 2, 'SZL', 'Lilangeni');
INSERT INTO iso_4217 VALUES (752, 2, 'SEK', 'Swedish Krona');
INSERT INTO iso_4217 VALUES (756, 2, 'CHF', 'Swiss Franc');
INSERT INTO iso_4217 VALUES (760, 2, 'SYP', 'Syrian Pound');
INSERT INTO iso_4217 VALUES (764, 2, 'THB', 'Baht');
INSERT INTO iso_4217 VALUES (776, 2, 'TOP', 'Pa’anga');
INSERT INTO iso_4217 VALUES (780, 2, 'TTD', 'Trinidad and Tobago Dollar');
INSERT INTO iso_4217 VALUES (784, 2, 'AED', 'UAE Dirham');
INSERT INTO iso_4217 VALUES (788, 3, 'TND', 'Tunisian Dinar');
INSERT INTO iso_4217 VALUES (800, 0, 'UGX', 'Uganda Shilling');
INSERT INTO iso_4217 VALUES (807, 2, 'MKD', 'Denar');
INSERT INTO iso_4217 VALUES (818, 2, 'EGP', 'Egyptian Pound');
INSERT INTO iso_4217 VALUES (826, 2, 'GBP', 'Pound Sterling');
INSERT INTO iso_4217 VALUES (834, 2, 'TZS', 'Tanzanian Shilling');
INSERT INTO iso_4217 VALUES (840, 2, 'USD', 'US Dollar');
INSERT INTO iso_4217 VALUES (858, 2, 'UYU', 'Peso Uruguayo');
INSERT INTO iso_4217 VALUES (860, 2, 'UZS', 'Uzbekistan Sum');
INSERT INTO iso_4217 VALUES (882, 2, 'WST', 'Tala');
INSERT INTO iso_4217 VALUES (886, 2, 'YER', 'Yemeni Rial');
INSERT INTO iso_4217 VALUES (901, 2, 'TWD', 'New Taiwan Dollar');
INSERT INTO iso_4217 VALUES (931, 2, 'CUC', 'Peso Convertible');
INSERT INTO iso_4217 VALUES (932, 2, 'ZWL', 'Zimbabwe Dollar');
INSERT INTO iso_4217 VALUES (934, 2, 'TMT', 'Turkmenistan New Manat');
INSERT INTO iso_4217 VALUES (936, 2, 'GHS', 'Ghana Cedi');
INSERT INTO iso_4217 VALUES (937, 2, 'VEF', 'Bolivar');
INSERT INTO iso_4217 VALUES (938, 2, 'SDG', 'Sudanese Pound');
INSERT INTO iso_4217 VALUES (940, 0, 'UYI', 'Uruguay Peso en Unidades Indexadas (URUIURUI)');
INSERT INTO iso_4217 VALUES (941, 2, 'RSD', 'Serbian Dinar');
INSERT INTO iso_4217 VALUES (943, 2, 'MZN', 'Mozambique Metical');
INSERT INTO iso_4217 VALUES (944, 2, 'AZN', 'Azerbaijanian Manat');
INSERT INTO iso_4217 VALUES (946, 2, 'RON', 'New Romanian Leu');
INSERT INTO iso_4217 VALUES (947, 2, 'CHE', 'WIR Euro');
INSERT INTO iso_4217 VALUES (948, 2, 'CHW', 'WIR Franc');
INSERT INTO iso_4217 VALUES (949, 2, 'TRY', 'Turkish Lira');
INSERT INTO iso_4217 VALUES (950, 0, 'XAF', 'CFA Franc BEAC');
INSERT INTO iso_4217 VALUES (951, 2, 'XCD', 'East Caribbean Dollar');
INSERT INTO iso_4217 VALUES (952, 0, 'XOF', 'CFA Franc BCEAO');
INSERT INTO iso_4217 VALUES (953, 0, 'XPF', 'CFP Franc');
INSERT INTO iso_4217 VALUES (967, 2, 'ZMW', 'Zambian Kwacha');
INSERT INTO iso_4217 VALUES (968, 2, 'SRD', 'Surinam Dollar');
INSERT INTO iso_4217 VALUES (969, 2, 'MGA', 'Malagasy Ariary');
INSERT INTO iso_4217 VALUES (970, 2, 'COU', 'Unidad de Valor Real');
INSERT INTO iso_4217 VALUES (971, 2, 'AFN', 'Afghani');
INSERT INTO iso_4217 VALUES (972, 2, 'TJS', 'Somoni');
INSERT INTO iso_4217 VALUES (973, 2, 'AOA', 'Kwanza');
INSERT INTO iso_4217 VALUES (974, 0, 'BYR', 'Belarussian Ruble');
INSERT INTO iso_4217 VALUES (975, 2, 'BGN', 'Bulgarian Lev');
INSERT INTO iso_4217 VALUES (976, 2, 'CDF', 'Congolese Franc');
INSERT INTO iso_4217 VALUES (977, 2, 'BAM', 'Convertible Mark');
INSERT INTO iso_4217 VALUES (978, 2, 'EUR', 'Euro');
INSERT INTO iso_4217 VALUES (979, 2, 'MXV', 'Mexican Unidad de Inversion (UDI)');
INSERT INTO iso_4217 VALUES (980, 2, 'UAH', 'Hryvnia');
INSERT INTO iso_4217 VALUES (981, 2, 'GEL', 'Lari');
INSERT INTO iso_4217 VALUES (984, 2, 'BOV', 'Mvdol');
INSERT INTO iso_4217 VALUES (985, 2, 'PLN', 'Zloty');
INSERT INTO iso_4217 VALUES (986, 2, 'BRL', 'Brazilian Real');
INSERT INTO iso_4217 VALUES (990, 4, 'CLF', 'Unidad de Fomento');
INSERT INTO iso_4217 VALUES (997, 2, 'USN', 'US Dollar (Next day)');

-- +goose Down
DROP VIEW v_account_position;
DROP TABLE account_position;
DROP VIEW v_contract;
DROP TABLE contract;
DROP TABLE exchange;
DROP TABLE symbol;
DROP TABLE security_type;
DROP TABLE account_amount;
DROP VIEW v_account_snapshot_latest;
DROP TABLE account_snapshot;
DROP TABLE account;
DROP TABLE account_type;
DROP TABLE iso_4217;
DROP TYPE monetary;
