#include "lock.pb.h"

/* Versão compatível com Nanopb 0.4.x */
const pb_msgdesc_t lock_SensorEvent_msg = {
    0, // Reservado
    (const uint32_t[]){
        PB_WT_STRING << 3 | 1, // field 1: device_id
        PB_WT_32BIT << 3 | 2,  // field 2: vibration_intensity
        PB_WT_STRING << 3 | 3, // field 3: entry_method
        PB_WT_VARINT << 3 | 4, // field 4: success
        0                      // Sentinel
    },
    (const uint16_t[]){
        offsetof(lock_SensorEvent, device_id),
        offsetof(lock_SensorEvent, vibration_intensity),
        offsetof(lock_SensorEvent, entry_method),
        offsetof(lock_SensorEvent, success),
        0
    },
    NULL, // Sem sub-mensagens
    NULL, // Sem valores default
    NULL, // Sem extensões
    sizeof(lock_SensorEvent),
    4,    // Número de campos
    0     // Descritores extra
};