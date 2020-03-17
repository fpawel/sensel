--[[

Функция product определяет тип ЧЕ:
    product(name, measurements...)
        name - строка, наименование типа ЧЭ
        measurements - список измерений, которые должны быть выполнены для данного типа ЧЭ.
            Объекты в списке measurements разделены запятыми и создаются в результате вызова функции measure.

Функция measure определяет измерение параметра ЧЭ и добавляет колонку в итоговую таблицу обмера:
    measure(name, gas, duration, I, U, calc)
        name - строка, наименование измерения, используемое при расчёте
        gas - целое число, номер газового клапана, который будет открыт перед измерением
        duration - строка, длительность продувки газа.
            Строка длительности - это знаковая последовательность десятичных чисел,
            каждое из которых содержит необязательную дробь и суффикс единицы времени,
            например "300ms", "-1.5h" или "2h45m".
            Допустимые единицы времени:
                "ns", "us" (или "µs"), "ms", "s", "m", "h".
        I - устанавливаемый стендом рабочий ток планки, А
        U - внешнее напряжение планки, В
        calc(x, prev) - функция, возвращая выводимое в итоговую таблицу значение.
            обект x (первый аргумент функции calc) содержит:
                поля, значения которых получены при измерении name:
                    x.U - измеренное напряжения ЧЭ
                    x.Q - измеренный расхода газа
                    x.I - измеренный ток
                    x.T - измеренная температура
                    x.C - концентрация ПГС
                метод x:Measure(prevName) - функцию, возвращающую результат y передыдущего измереня:
                    prevName - наименование предыдущего измерения
                    объект y, возвращаемый функцией x:Measure(prevName), представляет собой
                    результат передыдущего измерения prevName и содержит поля:
                        y.U, y.Q, y.I, y.T, y.C - аналогично объекту x
                        y.Value - расчитанное значение из итоговой таблицы
--]]


-- ГКС платины
local gammaPlatinum = 0.00385

-- нижний концентрационный предел распространения пламени CH4, проценты объёмных долей
local lowerFlammabilityLimitCH4 = 4.4

product('СГГ-1',

        measure('R0', 1, "1m", 5, 10, function(x)
            local R0 = x.U / ( x.I * (1 + gammaPlatinum * x.T))
            return R0, R0 >= 6.7 and R0 <= 7.3
        end),

        column("Uр", function(x)
            return x.U, x.U >= 1.75 and x.U <= 2.05
        end),

        column("Tчэ", function(x)
            local Ur = x:Measure('Uр').Value
            local R0 = x:Measure('R0').Value
            return (Ur / (x.I * R0) - 1) / gammaPlatinum, true
        end),

        column("Tчэ20", function(x)
            local Tch = x:Measure('Tчэ').Value
            return Tch - x.T + 20, true
        end),

        column("B", function(x)
            local Tch = x:Measure('Tчэ').Value
            local B = (Tch - x.T) / (x.I * x.I)
            return B, B >= 35000 and B <= 38000
        end),

        measure("Uгс", 2, "1m30s", 105, 50, function(x)
            return x.U, true
        end),

        column("K", function(x)
            local Ugs = x.U
            local Ur = x:Measure('Uр').Value
            local K = (lowerFlammabilityLimitCH4 / 100) * (Ugs - Ur) / x.C
            return K, K >= 3 and K <= 6
        end),

        measure("D", 2, "1m", 105, 50, function(x)
            local Ugs = x:Measure('Uгс').Value
            local Ur = x:Measure('Uр').Value
            return (x.U - Ugs) / (Ugs - Ur), true
        end)

)

product('СТМ-10 СКДМ',
        measure("X5", 1, "1m", 10, 10, function(x)
            return x.U / 2 + x.I + x.Q + x.T, true
        end),
        measure("X6", 1, "1m", 10, 10, function(x)
            local x5 = x:Measure('X5')
            return x.U + x.I + x.Q + x.T + x5.Value + x5.I + x5.Q + x5.T, true
        end)
)

