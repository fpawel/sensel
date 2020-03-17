-- ГКС платины
local gammaPlatinum = 0.00385

-- нижний концентрационный предел распространения пламени CH4, проценты объёмных долей
local lowerFlammabilityLimitCH4 = 4.4

device('СГГ-1',
        product('измерительный',
                measure(
                        sample(1, "1m", 5, 10),
                        sample(2, "1m30s", 105, 50),
                        sample(2, "1m30s", 105, 50),
                        sample(2, "1m30s", 105, 50)
                ),
                function(U, I, T, C)
                    local R0 = U[1] / (I[1] * (1 + gammaPlatinum * T[1]))

                    local Ur = U[2]
                    local Tch = (Ur / (I[2] * R0) - 1) / gammaPlatinum
                    local Tch20 = Tch - T[2] + 20
                    local B = (Tch - T[2]) / (I[2] * I[2])

                    local Ugs = U[3]
                    local K = (lowerFlammabilityLimitCH4 / 100) * (Ugs - Ur) / C[3]
                    local D = (U[4] - Ugs) / (Ugs - Ur)

                    return columns(
                            column("R0", R0, R0 >= 6.7 and R0 <= 7.3),

                            column("Ur", Ur, Ur >= 1.75 and Ur <= 2.05),
                            column("Tch", Tch, true),
                            column("Tch20", Tch20, true),
                            column("B", B, B >= 35000 and B <= 38000),

                            column("Uгс", Ugs, true),
                            column("K", K, K >= 3 and K <= 6),
                            column("D", D, true)
                    )
                end
        )
)