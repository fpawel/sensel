export.Devices =  {
    ['СТМ-30'] = {
        ['измерительный'] = require 'stm30m',
        ['сравнительный'] = require 'stm30c'
    },
    ['СТГ'] = {
        ['измерительный'] = require 'sgg10m',
        ['сравнительный'] = require 'sgg10c'
    },
	['СТМ-10'] = {
        ['сравнительный'] = require 'stm10m'
    }
}