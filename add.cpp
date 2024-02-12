#include "arm_neon.h"

int vsum(int count, int *input) {
  int32x4_t sum = vdupq_n_s32(0);

  for (int i = 0; i < count; i += 4) {
    int32x4_t v = vld1q_s32(&input[i]);
    vaddq_s32(sum, v);
  }

  int32x2_t sum2 = vadd_s32(vget_low_s32(sum), vget_high_s32(sum));
  int32x2_t sum3 = vpadd_s32(sum2, sum2);
  int32_t sum4 = vget_lane_s32(sum3, 0);

  return sum4;
}

int sum(int count, int *input) {
  int result = 0;
  for (int i = 0; i < count; i++) {
    result += input[i];
  }
  return result;
}

int main() {
  int count = 4096;
  int input[count];
  for (int i = 0; i < count; i++) {
    input[i] = i;
  }

  vsum(count, input);
}
