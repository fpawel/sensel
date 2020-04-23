


-- нижний концентрационный предел распространения пламени CH4, проценты объёмных долей
local lowerFlammabilityLimitCH4 = 4.4

export.Devices =  {
    ['СТМ-10'] = {
        ['сравнительный'] = {
            Samples = {
                sample(1, "30s", 9, 10),
                sample(1, "2m", 160, 50)
            },
            Calculate = function(U, I, T, C)

                -- ГКС платины
                local gammaPlatinum = 0.00385

                local R0 = U[1] / (I[1] * (1 + gammaPlatinum * T[1]))

                local Ur = U[2]
                local Tch = (Ur / (I[2] * R0) - 1) / gammaPlatinum
                local Tch20 = Tch - T[2] + 20
                local B = (Tch - T[2]) / (I[2] * I[2])

                return {
                    {
                        Name = "R0",
                        Value = R0,
                        Ok = R0 >= 3.63 and R0 <= 3.88,
                        Precision = 3,
                    },
                    {
                        Name = "Ur",
                        Value = Ur,
                        Ok = Ur >= 1.55 and Ur <= 1.8,
                        Precision = 3,
                    },
                    {
                        Name = "B",
                        Value = B,
                        Ok = B >= 15500 and B <= 19500,
                        Precision = 4,
                    }
                }
            end
        }
    }
}