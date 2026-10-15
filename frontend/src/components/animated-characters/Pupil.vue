<template>
  <div
    :data-max-distance="maxDistance"
    class="pupil"
    :style="pupilStyle"
  />
</template>

<script setup>


const {
  size,
  maxDistance,
  pupilColor
} = defineProps({
  size: '12px',
  maxDistance: 5,
  pupilColor: 'black'
})

const pupilStyle = {
  width: size,
  height: size,
  borderRadius: '50%',
  backgroundColor: pupilColor,
  willChange: 'transform'
}
</script>
