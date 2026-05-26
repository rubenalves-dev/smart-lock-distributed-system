<template>
  <div class="users-container" style="padding: 20px; font-family: sans-serif;">
    <h2>Lista de Utilizadores RFID</h2>
    <p style="color: #666; margin-bottom: 20px;">
      Total de utilizadores registados no sistema.
    </p>

    <table style="width: 100%; border-collapse: collapse; text-align: left;">
      <thead>
        <tr style="background-color: #f4f4f4; border-bottom: 2px solid #ddd;">
          <th style="padding: 12px;">RFID UID</th>
          <th style="padding: 12px;">Nome</th>
          <th style="padding: 12px;">E-mail</th>
          <th style="padding: 12px;">Estado do Perfil</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="user in users" :key="user.rfid" style="border-bottom: 1px solid #ddd;">
          <td style="padding: 12px; font-family: monospace; font-weight: bold;">{{ user.rfid }}</td>
          <td style="padding: 12px;">{{ user.name || '---' }}</td>
          <td style="padding: 12px;">{{ user.email || '---' }}</td>
          <td style="padding: 12px;">
            <span v-if="!user.name || !user.email" 
                  style="background-color: #ffcccc; color: #cc0000; padding: 4px 8px; border-radius: 4px; font-size: 12px; font-weight: bold;">
              Incomplete profile
            </span>
            <span v-else 
                  style="background-color: #d4edda; color: #155724; padding: 4px 8px; border-radius: 4px; font-size: 12px;">
              Ativo
            </span>
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'

// Definição da estrutura de dados
interface User {
  rfid: string;
  name: string | null;
  email: string | null;
}

// Declaração única com tipo definido
const users = ref<User[]>([])

const fetchUsers = async () => {
  try {
    const response = await fetch('/api/users')
    if (!response.ok) throw new Error('Erro na rede')
    users.value = await response.json()
  } catch (err) {
    console.error("Erro ao carregar utilizadores:", err)
  }
}

onMounted(() => {
  fetchUsers()
})
</script>