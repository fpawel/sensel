return {
    Samples = {
        sample(1, "1m", 5, 10),
        sample(2, "1m30s", 105, 50),
        sample(2, "1m30s", 105, 50),
        sample(2, "1m30s", 105, 50)
    },
    Calculate =  function(U, I, T, C)

        local lowerFlammabilityLimitCH4 = 4.4

        -- ГКС платины
        local gammaPlatinum = 0.00385

        local R0 = U[1] / (I[1] * (1 + gammaPlatinum * T[1]))

        local Ur = U[2]
        local Tch = (Ur / (I[2] * R0) - 1) / gammaPlatinum
        local Tch20 = Tch - T[2] + 20
        local B = (Tch - T[2]) / (I[2] * I[2])

        local Ugs = U[3]
        local K = (lowerFlammabilityLimitCH4 / 100) * (Ugs - Ur) / C[3]
        local D = (U[4] - Ugs) / (Ugs - Ur)

        return {
            {
                Name = "R0",
                Value = R0,
                Ok = R0 >= 6.7 and R0 <= 7.3,
                Precision = 3,
            },
            {
                Name = "Ur",
                Value = Ur,
                Ok = Ur >= 1.75 and Ur <= 2.05,
                Precision = 3,
            },
            {
                Name = "Tch",
                Value = Tch,
                Ok = true,
                Precision = 2,
            },
            {
                Name = "Tch20",
                Value = Tch20,
                Ok = true,
                Precision = 2,
            },
            {
                Name = "B",
                Value = B,
                Ok = B >= 35000 and B <= 38000,
                Precision = 4,
            },
            {
                Name = "Uгс",
                Value = Ugs,
                Ok = true,
                Precision = 2,
            },
            {
                Name = "K",
                Value = K,
                Ok = K >= 3 and K <= 6,
                Precision = 3,
            },
            {
                Name = "D",
                Value = D,
                Ok = true,
                Precision = 3,
            },
        }
    end
}

