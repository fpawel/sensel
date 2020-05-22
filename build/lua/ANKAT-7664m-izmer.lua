return {
    Samples = {
        sample(1, "0m30s", 0.003, 10),
        sample(1, "2m00s", 0.106, 30),
        sample(3, "4m00s", 0.106, 30),
        sample(3, "1m00s", 0.106, 30)
    },
    Calculate = function (U, I, T, C)

        -- Нижний концентрационный предел распространения пламени CH4, [% объёмной доли]
        local LEL = 4.4

        -- ТКС платины [1/°C]
        local gPt = 0.00385

        -- Номинальное значение рабочего тока [А]
        local Inom = 0.106

        -- Номинальное значение температуры среды [°C]
        local Tnom = 20

        -- Холодное сопротивление ЧЭ, приведенное к 0 °C [Ом]
        local R0 = U[1] / (I[1] * (1 + gPt * T[1]))

        -- Действительное напряжение на ЧЭ на ГС1 и токе I2 в ходе обмера [В]
        local Ur = U[2]
        
        -- Температура ЧЭ на ГС1 и токе I2, приведенная к температуре среды 20 °C [°C]
        local Tch_nom = (Ur / (I[2] * R0) - 1) / gPt - T[2] + Tnom
        
        -- Терморезистивный коэффициент [°C/А2]
        local B = (Tch_nom - Tnom) / (I[2] * I[2])

        -- Рабочее напряжение на ЧЭ, приведенное к 20 °C и номинальному значению рабочего тока [В]
        local U20 = Inom * R0 * (1 + gPt * (Tnom + B * Inom * Inom))

        -- Действительное напряжение на ЧЭ на ГС2 и токе I3 в ходе обмера [В]
        local Ugs = U[3]
    
        -- Чувствительность ЧЭ в схеме моста Уинстона [мВ/НКПР]
        local K = 1000 * (LEL / (100 * C[3])) * (Ugs - Ur) / 2

        -- Стабильность чувствительности на ГС [%]
        local D = 100 * (U[4] - Ugs) / Ugs

        return {
            {
                Name = "R0",
                Value = R0,
                Ok = R0 >= 3.80 and R0 <= 4.10,
                Precision = 3,
            },
            {
                Name = "Uизм",
                Value = Ur,
                Ok = true,
                Precision = 3,
            },
            {
                Name = "Tch20",
                Value = Tch_nom,
                Ok = true,
                Precision = 0,
            },
            {
                Name = "B",
                Value = B,
                Ok = true,
                Precision = 0,
            },
            {
                Name = "U20",
                Value = U20,
                Ok = U20 >= 1.030 and U20 <= 1.095,
                Precision = 3,
            },
            {
                Name = "K",
                Value = K,
                Ok = K >= 2 and K <= 4,
                Precision = 2,
            },
            {
                Name = "D",
                Value = D,
                Ok = D >= -0.5 and D <= 3.0,
                Precision = 2,
            },
        }
    end
}