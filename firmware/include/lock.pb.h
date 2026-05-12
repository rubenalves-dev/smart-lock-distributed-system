#ifndef LOCK_PB_H
#define LOCK_PB_H

#include <pb.h>

#ifdef __cplusplus
extern "C" {
#endif

typedef struct {
    char device_id[32];
    float vibration_intensity;
    char entry_method[16];
    bool success;
} lock_SensorEvent;

/* Descritor da mensagem */
extern const pb_msgdesc_t lock_SensorEvent_msg;

#define lock_SensorEvent_init_default {"", 0, "", false}

#ifdef __cplusplus
}
#endif

#endif