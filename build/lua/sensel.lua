Product('СГГ-1')
Measurement("X1", 1, "1m", function (M, C, I, Q, T)
    return M.X1 / 2 + I.X1 + Q.X1 + T.X1, true
end)
Measurement("X2", 1, "1m", function (M, C, I, Q, T)
    return M.X2 + C.X1  + I.X1 + Q.X1 + T.X1 + I.X2 + Q.X2 + T.X2, true
end)

Product('СТМ-10 СКДМ')
Measurement("X5", 1, "1m", function (M, C, I, Q, T)
    return M.X5 / 2 + I.X5 + Q.X5 + T.X5, true
end)
Measurement("X6", 1, "1m", function (M, C, I, Q, T)
    return M.X6 + C.X5  + I.X5 + Q.X5 + T.X5 + I.X6 + Q.X6 + T.X6, true
end)