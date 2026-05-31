<template>
  <el-container style="height: 100vh">
    <el-aside :width="isCollapse ? '64px' : '220px'" style="transition: width 0.3s">
      <Sidebar :is-collapse="isCollapse" />
    </el-aside>
    <el-container>
      <el-header style="padding: 0; border-bottom: 1px solid #e4e7ed">
        <Header @toggle-collapse="isCollapse = !isCollapse" />
      </el-header>
      <el-main style="background: #f5f7fa">
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import Sidebar from './Sidebar.vue'
import Header from './Header.vue'

const isCollapse = ref(false)
</script>
